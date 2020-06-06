package aws

/*
INFO:	this module contains logic for managing aws cloudwatch
USAGE:	first get an aws session, then get a cloudwatch client, which you can use to interact with aws API
		client := aws.GetCloudWatchClient()
*/

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

/*GetCloudWatchClient returns a new cloudwatch client*/
func GetCloudWatchClient(s *session.Session) *cloudwatch.CloudWatch {
	return cloudwatch.New(s)
}

/*ListMetrics lists the chosen metric for a particular instance
usage: ec2.ListMetrics("CPUUtilization", "AWS/EC2", "i-010cb2e16a05e05c0")*/
func ListMetrics(svc *cloudwatch.CloudWatch, metric string, namespace string, instanceID string) {
	// Disable the alarm
	result, err := svc.ListMetrics(&cloudwatch.ListMetricsInput{
		MetricName: aws.String(metric),
		Namespace:  aws.String(namespace),
		Dimensions: []*cloudwatch.DimensionFilter{
			&cloudwatch.DimensionFilter{
				Name:  aws.String("InstanceId"),
				Value: aws.String(instanceID),
			},
		},
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Metrics: ", result.Metrics)
}
