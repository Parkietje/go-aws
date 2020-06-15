package main

import (
	"fmt"
	"go-aws/m/v2/aws"
	"go-aws/m/v2/ingress"
	"go-aws/m/v2/loadbalancer"
	"log"
	"net/http"
)

func main() {

	//get an aws client
	sess, err := aws.GetSession()
	if err != nil {
		panic(err)
	}
	ec2Client := aws.GetEC2Client(sess)
	cloudwatchClient := aws.GetCloudWatchClient(sess)

	// Initialize the loadbalancer, this starts a worker pool with 1 worker
	loadbalancer.Initialize(ec2Client, cloudwatchClient, 1)

	// Start listening for post requests
	fmt.Println("Starting ingress server")
	ingress.Setup()
	http.HandleFunc("/", ingress.StyleTransfer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	// loadbalancer.TerminateAllWorkers(svc)
}
