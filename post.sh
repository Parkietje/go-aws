#!/bin/bash

#empty post
#curl -X POST http://localhost:8080

#multipart post:
curl --form "style=@style.png;type=image/png" --form "content=@content.png;type=image/png" --form size=50 --form iterations=1 http://localhost:8080
