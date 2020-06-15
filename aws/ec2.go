package aws

/*
INFO:	this module contains logic for managing aws ec2 resources
USAGE:	first get an aws session, then get an EC2 client, which you can use to interact with aws API
		svc := ec2.GetEC2Client()
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

/*GetEC2Client an aws EC2 client*/
func GetEC2Client(s *session.Session) (svc *ec2.EC2) {
	credentials, err := s.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println(credentials) //print the key stored in ~/.aws/credentials
	_ = credentials

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

//CheckInstanceState returns true when the instance with specified id has a public dns assigned
func CheckInstanceState(svc *ec2.EC2, ID string) bool {
	machineRunning := false
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(ID),
		},
	}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		if *result.Reservations[0].Instances[0].PublicDnsName != "" {
			machineRunning = true
		}
	}
	return machineRunning
}

/*GetPublicDNS returns the public dns of a given instance*/
func GetPublicDNS(svc *ec2.EC2, ID string) string {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(ID),
		},
	}
	result, err := svc.DescribeInstances(input)
	if err != nil {
		fmt.Println("Error", err)
	}
	return *result.Reservations[0].Instances[0].PublicDnsName
}

/*CreateInstance acquires a NEW resource (free tier use image "ami-085925f297f89fce1" and instance "t2.micro" )*/
func CreateInstance(svc *ec2.EC2, imageID string, instanceType string) *ec2.Instance {
	runResult, err := svc.RunInstances(&ec2.RunInstancesInput{
		ImageId:      aws.String(imageID),
		InstanceType: aws.String(instanceType),
		KeyName:      aws.String("go-aws"),
		MinCount:     aws.Int64(1),
		MaxCount:     aws.Int64(1),
	})

	if err != nil {
		log.Fatal(err)
	}

	// fmt.Println("Created instance", *runResult.Instances[0].InstanceId)
	// fmt.Println("Successfully created instance")

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
	}

	return runResult.Instances[0]
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

/*TerminateInstance terminates and releases the machine immediately*/
func TerminateInstance(svc *ec2.EC2, id string) {
	input := &ec2.TerminateInstancesInput{
		DryRun: aws.Bool(false),
		InstanceIds: []*string{
			aws.String(id),
		},
	}
	result, err := svc.TerminateInstances(input)
	if err != nil {
		fmt.Println("Error", err)
	} else {
		fmt.Println("Successfully terminated", result)
	}
}

/*MonitorInstance usage: state = "ON" or "OFF"
see: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-cloudwatch.html*/
func MonitorInstance(svc *ec2.EC2, state string, id string) {
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
