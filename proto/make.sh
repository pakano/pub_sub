#!/bin/bash

protoc --go_out=. *.proto
#protoc --go-grpc_out=require_unimplemented_servers=false:. *.proto
protoc --go-grpc_out=. *.proto