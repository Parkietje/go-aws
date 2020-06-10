package loadbalancer

import (
	"fmt"
	"sync"
	"time"

	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/ssh"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Worker struct that keeps track of the instance and status of the machine
type worker struct {
	instance *ec2.Instance
	active   bool
}

// Slice containing all active worker instances
var (
	workers          []worker
	loadbalancer_svc *ec2.EC2
	cloudwatch_svc   *cloudwatch.CloudWatch
)

var roundRobinIndex int = 0

func Initialize(svc *ec2.EC2, svc_cloudwatch *cloudwatch.CloudWatch, workerCount int) {

	// Make svc global
	loadbalancer_svc = svc
	cloudwatch_svc = svc_cloudwatch

	// Setup wait group
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		// Increment the wait group
		wg.Add(1)

		go func() {
			// Add a worker to the pool
			addWorker()
			// Decrement wait group
			wg.Done()
		}()
	}

	// Wait for all workers to initialize
	wg.Wait()

	go elasticScaling()

}

func addWorker() {
	// Request an ubuntu ami on a t2.micro machine type
	Inst := aws_helper.CreateInstance(loadbalancer_svc, "ami-068663a3c619dd892", "t2.micro")

	// Install the application on the instance over ssh
	ssh.InitializeWorker(loadbalancer_svc, *Inst.InstanceId)
	workers = append(workers, worker{
		instance: Inst,
		active:   true,
	})
}

func RunApplication(folder string) {
	// Round Robin Scheduling
	machine := workers[roundRobinIndex]

	// Increment the round robin index to cycle through the workers
	roundRobinIndex++
	if roundRobinIndex == len(workers) {
		roundRobinIndex = 0
	}

	fmt.Println("Scheduling request with folder id", folder, "on worker", *machine.instance.InstanceId)

	// Run the application via ssh
	ssh.RunApplication(loadbalancer_svc, *machine.instance.InstanceId, folder)

	fmt.Println("Application finished with folder id", folder)
}

func elasticScaling() {

	// Parameters
	scaleUpThres := 75.0
	scaleDownThres := 45.0
	maxWorkers := 5
	minWorkers := 1

	cumulativeCpu := 0.0
	aveCpu := 0.0

	// Endless loop
	for {
		cumulativeCpu = 0

		// Loop over all workers and collect cpu usage
		for _, machine := range workers {
			cumulativeCpu += aws_helper.GetCPUstats(cloudwatch_svc, *machine.instance.InstanceId)
		}

		aveCpu = cumulativeCpu / float64(len(workers))

		fmt.Println("Average worker CPU usage is", aveCpu)
		if aveCpu >= scaleUpThres && len(workers) < maxWorkers {
			// Add another worker to the pool
			go addWorker()
		}

		if aveCpu <= scaleDownThres && len(workers) > minWorkers {
			fmt.Println("Removing last worker", workers[len(workers)-1].instance.InstanceId, "from pool")
			// Remove last worker from pool
			workers = workers[:len(workers)-1]

			// TODO: wait for applications on worker to finish and terminate the instance
		}

		// Wait 1 minute
		time.Sleep(60 * time.Second)
	}

}

func TerminateAllWorkers(svc *ec2.EC2) {

	// Terminate all workers in the workers slice
	for index, machine := range workers {
		// Terminate the instance
		aws_helper.TerminateInstance(svc, *machine.instance.InstanceId)
		workers[index].active = false
	}

	fmt.Println("Successfully terminated", len(workers), "machine(s)")
}
