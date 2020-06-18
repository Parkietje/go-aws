#!/bin/bash

#empty post
#curl -X POST http://localhost:8080

#multipart post:
curl --form "style=@style.jpg;type=image/jpg" --form "content=@content.jpg;type=image/jpg" --form size=50 --form iterations=1 http://localhost:8080
