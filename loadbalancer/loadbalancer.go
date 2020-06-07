package loadbalancer

import (
	"fmt"

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

func Initialize(svc *ec2.EC2) {

	loadbalancer_svc = svc

	// Request an ubuntu ami on a t2.micro machine type
	Inst := aws_helper.CreateInstance(svc, "ami-07c1207a9d40bc3bd", "t2.micro")

	// Install the application on the instance over ssh
	ssh.InitializeWorker(svc, *Inst.InstanceId)
	workers = append(workers, worker{
		instance: Inst,
		active:   true,
	})

}

func RunApplication(folder string) {
	// TODO: round robin scheduling
	// Pick a worker to run on, for now pick the first one
	machine := workers[0]

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
