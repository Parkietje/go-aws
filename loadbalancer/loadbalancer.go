package loadbalancer

import (
	"fmt"
	"sync"

	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/ssh"

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
)

var roundRobinIndex int = 0

func Initialize(svc *ec2.EC2, workerCount int) {

	// Make svc global
	loadbalancer_svc = svc

	// Setup wait group
	var wg sync.WaitGroup

	for i := 0; i < workerCount; i++ {
		// Increment the wait group
		wg.Add(1)

		go func() {
			// Request an ubuntu ami on a t2.micro machine type
			Inst := aws_helper.CreateInstance(svc, "ami-068663a3c619dd892", "t2.micro")

			// Install the application on the instance over ssh
			ssh.InitializeWorker(svc, *Inst.InstanceId)
			workers = append(workers, worker{
				instance: Inst,
				active:   true,
			})

			// Decrement wait group
			wg.Done()
		}()
	}

	// Wait for all workers to initialize
	wg.Wait()

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
