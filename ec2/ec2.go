package ec2

/*
INFO:	this module contains logic for managing aws ec2 resources
USAGE:	first get an aws client, which you can use to interact with aws API
		svc := ec2.GetClient()
		ec2.GetInstanceConsoleOutput(svc, "instanceID")
*/

import (
	"encoding/base64"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

/*GetClient starts a new session and returs an aws client*/
func GetClient() (svc *ec2.EC2) {
	//aws will look for credentials and config specified by environment variables
	s, err := session.NewSession(nil)

	if err != nil {
		log.Fatal(err)
	}

	credentials, err := s.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(credentials) //print the key stored in ~/.aws/credentials

	// Create new EC2 client
	return ec2.New(s)
}

/*DescribeInstances get detailed info about all instances*/
func DescribeInstances(svc *ec2.EC2) {
	result, err := svc.DescribeInstances(nil)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Describe all instances: ", result)
	}
}

/*CreateInstance acquires a NEW resource (free tier use image "ami-085925f297f89fce1" and instance "t2.micro" )*/
func CreateInstance(svc *ec2.EC2, imageID string, instanceType string) {
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: aws.String(instanceType),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})

	if err != nil {
		log.Fatal(err)
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

/*StartInstance can be used to start a particular instance*/
func StartInstance(svc *ec2.EC2, id string) {
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
}

/*StopInstance can be used to stop a particular instance*/
func StopInstance(svc *ec2.EC2, id string) {
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

/*RebootInstance can be used when we need to escape faulty state or crash*/
func RebootInstance(svc *ec2.EC2, id string) {
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

/*see: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch.html*/
func monitorInstance(svc *ec2.EC2, state string, id string) {
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

/*GetInstanceConsoleOutput gets the aggregated console output of a particular instance*/
func GetInstanceConsoleOutput(svc *ec2.EC2, id string) {
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
