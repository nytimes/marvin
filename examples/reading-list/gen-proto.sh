#!/bin/sh

go get -u github.com/NYTimes/openapi2proto/cmd/openapi2proto

openapi2proto -spec reading-list.yaml > reading-list.proto;

# for our code
protoc --go_out=. reading-list.proto;
