package main

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
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

	//reboot instance with id
	//rebootEC2(svc, "i-010fd3e08f862fed3")

	//monitor instance with id
	//monitorEC2(svc, "OFF", "i-010fd3e08f862fed3")

	//get console output for instance with id
	//getConsoleOutputEC2(svc, "i-010fd3e08f862fed3")

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
				fmt.Println("Successfully started", result.StartingInstances)
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
				fmt.Println("Successfully stopped", result.StoppingInstances)
			}
		} else {
			fmt.Println("Error", err)
		}
	}
}

func rebootEC2(svc *ec2.EC2, id string) {
	// We set DryRun to true to check to see if the instance exists and we have the
	// necessary permissions to monitor the instance.
	input := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			aws.String(id),
		},
		DryRun: aws.Bool(true),
	}
	result, err := svc.RebootInstances(input)
	awsErr, ok := err.(awserr.Error)

	// If the error code is `DryRunOperation` it means we have the necessary
	// permissions to Start this instance
	if ok && awsErr.Code() == "DryRunOperation" {
		// Let's now set dry run to be false. This will allow us to reboot the instances
		input.DryRun = aws.Bool(false)
		result, err = svc.RebootInstances(input)
		if err != nil {
			fmt.Println("Error", err)
		} else {
			fmt.Println("Successfully rebooted", result)
		}
	} else { // This could be due to a lack of permissions
		fmt.Println("Error", err)
	}
}

func monitorEC2(svc *ec2.EC2, state string, id string) {
	// Turn monitoring on
	if state == "ON" {
		// We set DryRun to true to check to see if the instance exists and we have the
		// necessary permissions to monitor the instance.
		input := &ec2.MonitorInstancesInput{
			InstanceIds: []*string{
				aws.String(id),
			},
			DryRun: aws.Bool(true),
		}
		result, err := svc.MonitorInstances(input)
		awsErr, ok := err.(awserr.Error)

		// If the error code is `DryRunOperation` it means we have the necessary
		// permissions to monitor this instance
		if ok && awsErr.Code() == "DryRunOperation" {
			// Let's now set dry run to be false. This will allow us to turn monitoring on
			input.DryRun = aws.Bool(false)
			result, err = svc.MonitorInstances(input)
			if err != nil {
				fmt.Println("Error", err)
			} else {
				fmt.Println("Successful monitoring ON", result.InstanceMonitorings)
			}
		} else {
			// This could be due to a lack of permissions
			fmt.Println("Error", err)
		}
	} else if state == "OFF" { // Turn monitoring off
		input := &ec2.UnmonitorInstancesInput{
			InstanceIds: []*string{
				aws.String(id),
			},
			DryRun: aws.Bool(true),
		}
		result, err := svc.UnmonitorInstances(input)
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == "DryRunOperation" {
			input.DryRun = aws.Bool(false)
			result, err = svc.UnmonitorInstances(input)
			if err != nil {
				fmt.Println("Error", err)
			} else {
				fmt.Println("Successful monitoring OFF", result.InstanceMonitorings)
			}
		} else {
			fmt.Println("Error", err)
		}
	}
}

func decodeOutputb64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Println("Error", err)
		return ""
	}

	return string(decoded)
}

func getConsoleOutputEC2(svc *ec2.EC2, id string) {
	// Call EC2 GetConsoleOutput API on the given instance according
	//   https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.GetConsoleOutput
	input := ec2.GetConsoleOutputInput{InstanceId: aws.String(id)}
	json, err := svc.GetConsoleOutput(&input)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("console output:")
		fmt.Println(decodeOutputb64(*json.Output))
	}
}
