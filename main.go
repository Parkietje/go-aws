package main

import (
	aws_helper "go-aws/m/v2/aws"
	"go-aws/m/v2/loadbalancer"
)

func main() {

	//get an aws client
	sess, err := aws_helper.GetSession()
	if err != nil {
		panic(err)
	}
	svc := aws_helper.GetEC2Client(sess)

	// Initialize the loadbalancer, this starts a worker pool with 1 worker
	loadbalancer.Initialize(svc)

	loadbalancer.RunApplication(svc)

	// Start listening for post requests
	// ingress.Setup()
	// http.HandleFunc("/", ingress.StyleTransfer)
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }

	// loadbalancer.TerminateAllWorkers(svc)
}
