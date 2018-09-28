#!/bin/sh

go get -u github.com/NYTimes/openapi2proto/cmd/openapi2proto

openapi2proto -spec service.yaml > service.proto;

# for our code
protoc --go_out=. service.proto;
