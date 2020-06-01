package main

import (
	"go-aws/m/v2/ec2"
)

func main() {
	//get an aws client
	svc := ec2.GetClient()

	//describe all instances
	ec2.DescribeInstances(svc)

	//ec2.StopInstance(svc, "i-010fd3e08f862fed3")
}
