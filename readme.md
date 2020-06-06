# go-aws

aws management tool

# run

`$ go run main.go`

# config

add aws ssh key in `./keys/yourkey.pem`

<!-- TODO: very unsecure, only the loadbalancers IP should be whitelisted automatically -->
allow ssh connections from any ip in the defaut security group


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