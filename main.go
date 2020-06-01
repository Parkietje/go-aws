package main

import (
	"fmt"
	"go-aws/m/v2/ec2"
	"log"
)

func main() {
	//get an aws client
	svc, err := ec2.GetClient()
	if err != nil {
		log.Fatal(err)
	}

	//describe all instances
	result, err := ec2.DescribeInstances(svc)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Success", result)
	}

}
