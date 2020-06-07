package main

import (
	"go-aws/m/v2/ingress"

	"log"
	"net/http"
	ec2_helper "go-aws/m/v2/ec2"
	"go-aws/m/v2/ssh"
)

func main() {
	//get an aws client
	svc := ec2_helper.GetClient()

	//describe all instances
	// ec2_helper.DescribeInstances(svc)
	// Start listening for post requests
	ingress.Setup()
	http.HandleFunc("/", ingress.StyleTransfer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	// TODO: extract this to a loadbalancer package and automate spinning up new instances
	// Request an ubuntu ami on a t2.micro machine type
	Inst := ec2_helper.CreateInstance(svc, "ami-07c1207a9d40bc3bd", "t2.micro")
	// Install the application on the instance over ssh
	ssh.InitializeWorker(svc, *Inst.InstanceId)

	// Stopping an instance does not realease the machine, use terminate instead
	//ec2.StopInstance(svc, "i-010fd3e08f862fed3")
	ec2_helper.TerminateInstance(svc, *Inst.InstanceId)
}
