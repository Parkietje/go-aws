package main

import (
	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/ssh"
)

func main() {
	// Start listening for post requests
	// ingress.Setup()
	// http.HandleFunc("/", ingress.StyleTransfer)
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }

	//get an aws client
	sess, err := aws_helper.GetSession()
	if err != nil {
		panic(err)
	}
	svc := aws_helper.GetEC2Client(sess)

	//describe all instances
	// aws_helper.DescribeInstances(svc)

	// TODO: extract this to a loadbalancer package and automate spinning up new instances
	// Request an ubuntu ami on a t2.micro machine type
	Inst := aws_helper.CreateInstance(svc, "ami-07c1207a9d40bc3bd", "t2.micro")
	// Install the application on the instance over ssh
	ssh.InitializeWorker(svc, *Inst.InstanceId)

	// Stopping an instance does not realease the machine, use terminate instead
	//aws_helper.StopInstance(svc, "i-010fd3e08f862fed3")
	aws_helper.TerminateInstance(svc, *Inst.InstanceId)
}
