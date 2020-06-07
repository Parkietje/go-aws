package aws

/*
INFO:	this module contains logic for managing aws cloudwatch
USAGE:	first get an aws session, then get a cloudwatch client, which you can use to interact with aws API
		client := aws.GetCloudWatchClient()
*/

import (
	"fmt"
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
func GetCPUstats(svc *cloudwatch.CloudWatch, id string) {
	var input cloudwatch.GetMetricStatisticsInput
	now := time.Now()
	start := now.Add(time.Duration(-5) * time.Minute) //5 minutes ago
	input.EndTime = aws.Time(now)
	input.StartTime = aws.Time(start)
	input.MetricName = aws.String("CPUUtilization")
	input.Namespace = aws.String("AWS/EC2")
	dimension := cloudwatch.Dimension{
		Name:  aws.String("InstanceId"),
		Value: aws.String(id),
	}
	input.Dimensions = []*cloudwatch.Dimension{&dimension}
	input.Period = aws.Int64(60) //1 min
	input.Statistics = []*string{aws.String("Average")}

	result, err := svc.GetMetricStatistics(&input)
	if err != nil {
		print(err.Error())
	}
	fmt.Println(result.GoString())
}
