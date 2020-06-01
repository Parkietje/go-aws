package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	//aws will look for credentials and config specified by environment variables
	s, err := session.NewSession(nil)

	if err != nil {
		log.Fatal(err)
	}

	creds, err := s.Config.Credentials.Get()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(creds) //print the key stored in ~/.aws/credentials

	// Create new EC2 client
	svc := ec2.New(s)

	//create new ec2 instance
	//createEC2(svc)

	//start/stop instance with ID
	//startStopEC2(svc, "STOP", "i-010fd3e08f862fed3")

	// Call to get detailed information on each instance
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Success", result)
	}
}

func createEC2(svc *ec2.EC2) {
	// Specify the details of the instance that you want to create.
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		// An Amazon Ubuntu AMI ID for t2.micro
		ImageId:      aws.String("ami-085925f297f89fce1"),
		InstanceType: aws.String("t2.micro"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})

	if err != nil {
		fmt.Println("Could not create instance", err)
		return
	}

	fmt.Println("Created instance", *runResult.Instances[0].InstanceId)

	// Add tags to the created instance
	_, errtag := svc.CreateTags(&ec2.CreateTagsInput{
		Resources: []*string{runResult.Instances[0].InstanceId},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("go-aws"),
			},
		},
	})
	if errtag != nil {
		log.Println("Could not create tags for instance", runResult.Instances[0].InstanceId, errtag)
		return
	}

	fmt.Println("Successfully tagged instance")
}

func startStopEC2(svc *ec2.EC2, state string, id string) {
	// Turn monitoring on
	if state == "START" {
		// We set DryRun to true to check to see if the instance exists and we have the
		// necessary permissions to monitor the instance.
		input := &ec2.StartInstancesInput{
			InstanceIds: []*string{
				aws.String(id),
			},
			DryRun: aws.Bool(true),
		}
		result, err := svc.StartInstances(input)
		awsErr, ok := err.(awserr.Error)

		// If the error code is `DryRunOperation` it means we have the necessary
		// permissions to Start this instance
		if ok && awsErr.Code() == "DryRunOperation" {
			// Let's now set dry run to be false. This will allow us to start the instances
			input.DryRun = aws.Bool(false)
			result, err = svc.StartInstances(input)
			if err != nil {
				fmt.Println("Error", err)
			} else {
				fmt.Println("Success", result.StartingInstances)
			}
		} else { // This could be due to a lack of permissions
			fmt.Println("Error", err)
		}
	} else if state == "STOP" { // Turn instances off
		input := &ec2.StopInstancesInput{
			InstanceIds: []*string{
				aws.String(id),
			},
			DryRun: aws.Bool(true),
		}
		result, err := svc.StopInstances(input)
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "DryRunOperation" {
			input.DryRun = aws.Bool(false)
			result, err = svc.StopInstances(input)
			if err != nil {
				fmt.Println("Error", err)
			} else {
				fmt.Println("Success", result.StoppingInstances)
			}
		} else {
			fmt.Println("Error", err)
		}
	}
}
