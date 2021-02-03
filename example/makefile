SHELL=/bin/bash

.PHONY:all
all:clean build

build:
	protoc -I=. --go_out=plugins=grpc:. --go-http_out=. --swagger_out=. ./pb/demo.proto

clean:
	rm -rf ./pb/*.go ./pb/*.yaml