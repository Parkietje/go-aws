package loadbalancer

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/ssh"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// Worker struct that keeps track of the instance and status of the machine
type worker struct {
	instance  *ec2.Instance
	active    bool
	ip        string
	heartbeat int64
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
	go MonitorHeartbeats()

}

func addWorker() {
	// Request an ubuntu ami on a t2.micro machine type
	Inst := aws_helper.CreateInstance(loadbalancer_svc, "ami-068663a3c619dd892", "t2.micro")

	// Install the application on the instance over ssh
	ssh.InitializeWorker(loadbalancer_svc, *Inst.InstanceId)
	publicDns := aws_helper.RetrivePublicDns(loadbalancer_svc, *Inst.InstanceId)
	workers = append(workers, worker{
		instance:  Inst,
		active:    true,
		ip:        strings.Replace(strings.Split(publicDns, ".")[0][4:], "-", ".", -1), // TODO: clean this up
		heartbeat: time.Now().Unix(),
	})
}

func RunApplication(folder string) {
	// Round Robin Scheduling
	machine := workers[roundRobinIndex]

	// Only run on active workers
	for machine.active == false {
		// Increment the round robin index to cycle through the workers
		roundRobinIndex++
		if roundRobinIndex == len(workers) {
			roundRobinIndex = 0
		}

		machine = workers[roundRobinIndex]
	}

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
	scaleUpThres := 50.0
	scaleDownThres := 30.0
	maxWorkers := 5
	minWorkers := 1

	cumulativeCpu := 0.0
	aveCpu := 0.0

	// Endless loop
	for {
		cumulativeCpu = 0

		// Loop over all workers and collect cpu usage
		var activeWorkerCount int
		for _, machine := range workers {
			// cumulativeCpu += aws_helper.GetCPUstats(cloudwatch_svc, *machine.instance.InstanceId)
			if machine.active {
				cumulativeCpu += ssh.GetCpuUtilization(loadbalancer_svc, *machine.instance.InstanceId)
				activeWorkerCount++
			}
		}

		aveCpu = cumulativeCpu / float64(activeWorkerCount)

		fmt.Println("Average worker CPU usage is", aveCpu, "with a total of", activeWorkerCount, "active workers")
		if aveCpu >= scaleUpThres && len(workers) < maxWorkers {
			// Add another worker to the pool
			go addWorker()
		}

		if aveCpu <= scaleDownThres && len(workers) > minWorkers {
			fmt.Println("Removing last worker", *workers[len(workers)-1].instance.InstanceId, "from pool")

			workerToTerminate := workers[len(workers)-1]

			// Remove last worker from pool
			workers = workers[:len(workers)-1]

			go waitForApplicationsToFinishAndTerminate(workerToTerminate)
		}

		// Wait 30 seconds
		time.Sleep(30 * time.Second)
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

func waitForApplicationsToFinishAndTerminate(machine worker) {

	// Wait for applications to finish
	for ssh.CheckIfApplicationsAreRunning(loadbalancer_svc, *machine.instance.InstanceId) == true {
	}

	aws_helper.TerminateInstance(loadbalancer_svc, *machine.instance.InstanceId)

	fmt.Println("Worker", *machine.instance.InstanceId, "terminated")
}

func MonitorHeartbeats() {
	l, err := net.Listen("tcp4", ":8000")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func() {
			remoteIp := strings.Split(c.RemoteAddr().String(), ":")[0]

			var machine worker
			for _, m := range workers {
				if m.ip == remoteIp {
					machine = m
				}
			}

			if machine.instance.InstanceId != nil {
				fmt.Println("Received heartbeat from", *machine.instance.InstanceId)
				timestamp := time.Now().Unix()
				machine.heartbeat = timestamp

				go func() {
					time.Sleep(60 * time.Second)

					for _, m := range workers {
						if m.ip == remoteIp {
							machine = m
						}
					}

					if machine.heartbeat == timestamp {
						// No new heartbeat received

						// Set machine inactive
						for _, m := range workers {
							if m.ip == remoteIp {
								m.active = false
							}
						}

						// Reboot the machine
						fmt.Println("Rebooting failed worker", *machine.instance.InstanceId)
						// aws_helper.RebootInstance(loadbalancer_svc, *machine.instance.InstanceId) // Do this in production
						aws_helper.StartInstance(loadbalancer_svc, *machine.instance.InstanceId)

						machine.heartbeat = time.Now().Unix()

						// Set machine active
						for _, m := range workers {
							if m.ip == remoteIp {
								m.active = true
							}
						}
					}
				}()

			} else {
				fmt.Println("Received heartbeat from unknown machine")
			}
			c.Close()
		}()
	}
}
