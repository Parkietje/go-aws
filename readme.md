# go-aws

aws management tool

# run

`$ go run main.go`

# config


### UNIX:

add aws access key in `~/.aws/credentials` using header `[go-aws]`

add aws config in `~/.aws/config` using header `[go-aws]`

set environment variables for loading config files:

`$ export AWS_SDK_LOAD_CONFIG=true`

`$ export AWS_PROFILE=go-aws`

### Windows:

add aws access key in `%UserProfile%\.aws\credentials` using header `[go-aws]`

add aws config in `%UserProfile%\.aws\config` using header `[go-aws]`

set environment variables for loading config files:

`$ C:\> setx AWS_SDK_LOAD_CONFIG true`

`$ C:\> setx AWS_PROFILE go-aws`

# SSH into instances

import your public key with aws cli:

`$ aws ec2 import-key-pair --key-name "YOUR_KEY_NAME" --public-key-material file://path/to/keypair/my-key.pub`

an instance which is provisioned with your public key can be reached by ssh'ing with your private key:

`$ ssh -i path/to/keypair/my-key  USERNAME@INSTANCE_IP`

the default username depends on the AMI you chose (e.g. ubuntu default username = ubuntu)