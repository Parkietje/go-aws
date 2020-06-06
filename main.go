package main

import (
	ec2_helper "go-aws/m/v2/ec2"
	"go-aws/m/v2/ssh"
)

func main() {
	//get an aws client
	svc := ec2_helper.GetClient()

	//describe all instances
	ec2_helper.DescribeInstances(svc)

	Inst := ec2_helper.CreateInstance(svc, "ami-07c1207a9d40bc3bd", "t2.micro")

	//ec2.DescribeInstances(svc)

	ssh.InitializeWorker(svc, Inst)

	ec2_helper.TerminateInstance(svc, *Inst.InstanceId)

	//ec2.DescribeInstances(svc)

	//ec2.StopInstance(svc, "i-010fd3e08f862fed3")
}
