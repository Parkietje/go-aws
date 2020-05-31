package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var i int = 0

func main() {
	//aws will look for credentials and config specified by environment variables
	sess, err := session.NewSession(nil)

	if err != nil {
		log.Fatal(err)
	}

	creds, err := sess.Config.Credentials.Get()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(creds) //print the key stored in ~/.aws/credentials

	// Create new EC2 client
	ec2svc := ec2.New(sess)

	// Call to get detailed information on each instance
	result, err := ec2svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Success", result)
	}
}
