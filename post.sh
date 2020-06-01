#!/bin/bash

#empty post
curl -X POST http://localhost:8080

#multipart post:
#curl --form "fileupload=@my-file.txt;filename=desired-filename.txt" --form param1=value1 --form param2=value2 http://localhost:8080
