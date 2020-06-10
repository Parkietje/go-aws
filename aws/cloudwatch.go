package aws

/*
INFO:	this module contains logic for managing aws cloudwatch
USAGE:	first get an aws session, then get a cloudwatch client, which you can use to interact with aws API
		client := aws.GetCloudWatchClient()
*/

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

/*GetCloudWatchClient returns a new cloudwatch client*/
func GetCloudWatchClient(s *session.Session) *cloudwatch.CloudWatch {
	return cloudwatch.New(s)
}

/*GetCPUstats returns the 5 min average CPU Utilization for a particular instance*/
func GetCPUstats(svc *cloudwatch.CloudWatch, instanceID string) float64 {
	var input cloudwatch.GetMetricStatisticsInput
	now := time.Now()
	start := now.Add(time.Duration(-10) * time.Minute)
	input.EndTime = aws.Time(now)
	input.StartTime = aws.Time(start)
	input.MetricName = aws.String("CPUUtilization")
	input.Namespace = aws.String("AWS/EC2")
	dimension := cloudwatch.Dimension{
		Name:  aws.String("InstanceId"),
		Value: aws.String(instanceID),
	}
	input.Dimensions = []*cloudwatch.Dimension{&dimension}
	input.Period = aws.Int64(5 * 60)
	input.Statistics = []*string{aws.String("Average")}

	result, err := svc.GetMetricStatistics(&input)
	if err != nil {
		print(err.Error())
	}

	percentage := 0.0
	if result.Datapoints != nil {
		percentage = *result.Datapoints[len(result.Datapoints)-1].Average
	}

	return percentage
}
