#!/bin/bash

#environment variables
export AWS_SDK_LOAD_CONFIG=true 
export AWS_PROFILE=go-aws

#for development:
go run main.go #comment this in production

#for production first build:
#go build main.go

#and then execute the binary:
#./main