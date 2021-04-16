SHELL=/bin/bash

.PHONY:all
all: clean build_swagger

clean:
	rm -rf ./elfpuck/*.go

build_swagger:
	protoc -I=. -I=$(GOPATH)/src --go_out=plugins=grpc:. ./protoc-gen-swagger/options/swagger.proto
