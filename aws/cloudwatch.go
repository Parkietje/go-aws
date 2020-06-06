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

/*ListMetrics usage: ec2.ListMetrics("CPUUtilization", "AWS/EC2", "Name=InstanceId,Value=i-010cb2e16a05e05c0")*/
func ListMetrics(svc *cloudwatch.CloudWatch, metric string, namespace string, dimension string) {
	// Disable the alarm
	result, err := svc.ListMetrics(&cloudwatch.ListMetricsInput{
		MetricName: aws.String(metric),
		Namespace:  aws.String(namespace),
		Dimensions: []*cloudwatch.DimensionFilter{
			&cloudwatch.DimensionFilter{
				Name: aws.String(dimension),
			},
		},
	})
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Metrics: ", result.Metrics)
}
