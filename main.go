package main

import (
	"fmt"
	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/ingress"
	"go-aws/m/v2/loadbalancer"
	"log"
	"net/http"
)

func main() {

	//get an aws client
	sess, err := aws_helper.GetSession()
	if err != nil {
		panic(err)
	}
	svc := aws_helper.GetEC2Client(sess)
	svc_cloudwatch := aws_helper.GetCloudWatchClient(sess)

	// Initialize the loadbalancer, this starts a worker pool with 1 worker
	loadbalancer.Initialize(svc, svc_cloudwatch, 5)

	// Start listening for post requests
	fmt.Println("Starting ingress server")
	ingress.Setup()
	http.HandleFunc("/", ingress.StyleTransfer)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	// loadbalancer.TerminateAllWorkers(svc)
}
