package loadbalancer

import (
	"fmt"
	"go-aws/m/v2/aws"
	"strconv"
	"sync"
	"time"

	"go-aws/m/v2/logger"
	"go-aws/m/v2/ssh"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

/*Worker is a struct that keeps track of the instance and status of the machine*/
type worker struct {
	instance *ec2.Instance
	active   bool
}

var (
	workers          []worker                   // Slice with active worker instances
	ec2Client        *ec2.EC2                   // AWS client to access ec2 API
	cloudwatchClient *cloudwatch.CloudWatch     // AWS client for cloudwatch API
	roundRobinIndex  int                    = 0 // index to keep track of scheduling
)

const (
	defaultAMI          = "ami-068663a3c619dd892" // ubuntu AMI
	defaultInstanceType = "c5.large"              // free tier default instance type
)

/*Initialize the loadbalancer with the specified amount of workers*/
func Initialize(ec2 *ec2.EC2, cloudwatch *cloudwatch.CloudWatch, workerCount int) {
	ec2Client = ec2
	cloudwatchClient = cloudwatch
	// Setup wait group for each worker
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			addWorker(defaultAMI, defaultInstanceType)
			wg.Done()
		}()
	}
	// Wait for all workers to initialize
	wg.Wait()

	// Initialize the logger, this creates the csv file
	logger.Init()

	go elasticScaling()
}

func addWorker(AMI string, instanceType string) {
	Inst := aws.CreateInstance(ec2Client, AMI, instanceType)
	// Install the application on the instance over ssh
	ssh.InitializeWorker(ec2Client, *Inst.InstanceId)
	workers = append(workers, worker{
		instance: Inst,
		active:   true,
	})

	// Log the new number of workers
	workerCount := strconv.Itoa(len(workers)) // This has to be a string
	logger.Log(workerCount)
}

func elasticScaling() {
	// Parameters
	scaleUpThres := 50.0
	scaleDownThres := 30.0
	maxWorkers := 5
	minWorkers := 1

	sumCPU := 0.0
	avgCPU := 0.0

	for {
		sumCPU = 0
		// Loop over all workers and collect cpu usage
		for _, machine := range workers {
			res, err := ssh.GetCPUUtilization(ec2Client, *machine.instance.InstanceId)
			//TODO on error should we retry? or reboot?
			if err == nil {
				sumCPU += res
			}
		}
		avgCPU = sumCPU / float64(len(workers))
		fmt.Println("Average worker CPU usage is", avgCPU, "with a total of", len(workers), "workers")

		// Add another worker to the pool if CPU exceeds thresh and workers < max
		if avgCPU >= scaleUpThres && len(workers) < maxWorkers {
			go addWorker(defaultAMI, defaultInstanceType)
		}

		// Remove worker from pool if CPU below thresh and workers > min
		if avgCPU <= scaleDownThres && len(workers) > minWorkers {
			fmt.Println("Removing worker", *workers[len(workers)-1].instance.InstanceId, "from pool")
			workerToTerminate := workers[len(workers)-1]
			workers = workers[:len(workers)-1]
			go waitForApplicationsToFinishAndTerminate(workerToTerminate)
		}

		time.Sleep(30 * time.Second)
	}
}

func waitForApplicationsToFinishAndTerminate(machine worker) {
	for ssh.InstanceRunning(ec2Client, *machine.instance.InstanceId) == true {
	}
	aws.TerminateInstance(ec2Client, *machine.instance.InstanceId)
	fmt.Println("Worker", *machine.instance.InstanceId, "terminated")
}

/*RunApplication schedules a worker to process the contents of the given folder*/
func RunApplication(folder string, styleFile string, contentFile string) (err error) {

	//round robin scheduling
	machine := workers[roundRobinIndex]
	roundRobinIndex++
	if roundRobinIndex == len(workers) {
		roundRobinIndex = 0
	}

	fmt.Println("Scheduling request with folder id", folder, "on worker", *machine.instance.InstanceId)

	// Run the application via ssh
	err = ssh.RunApplication(ec2Client, *machine.instance.InstanceId, folder, styleFile, contentFile)
	return
}

/*TerminateAllWorkers shuts down all workers immediately*/
func TerminateAllWorkers(svc *ec2.EC2) {
	for index, machine := range workers {
		// Terminate the instance
		aws.TerminateInstance(svc, *machine.instance.InstanceId)
		workers[index].active = false
	}
	fmt.Println("Successfully terminated", len(workers), "machine(s)")
}
