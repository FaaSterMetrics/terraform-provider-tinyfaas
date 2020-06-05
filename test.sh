#!/bin/bash

# expects a tinyfaas instance running on localhost

go build -o terraform-provider-tinyfaas
#terraform init
tar -cvf blubb.tar -C blubb/ .
terraform apply
sleep 1
curl localhost:5683/blubb