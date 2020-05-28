package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func main() {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		//add access key with header [go-aws] in ~/.aws/credentials
		Credentials: credentials.NewSharedCredentials("", "go-aws"),
	})

	if err != nil {
		log.Fatal(err)
	}

	creds, err := sess.Config.Credentials.Get()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(creds) //print the key stored in ~/.aws/credentials
}
