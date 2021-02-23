## protoc-gen-go-http

## Installation
```shell script
go get -u github.com/gogo/protobuf/protoc-gen-gofast            #生成go代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-go-http     #生成grpc-http 代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-swagger     #生成swagger
```

## Generate
```shell script
protoc -I=. -I=${GOPATH}/src --gofast_out=plugins=grpc:. --http_out=. --swagger_out=. xxx.proto
```

## Usage
```go
package main

import (
	"context"
	demo "example/pb"
	"fmt"
	"google.golang.org/grpc"
	"github.com/elfpuck/grpc-http/elfHttp"
)

func main() {
	e := elfHttp.New()
	service := e.Service("demo")
	demo.RegisterDemoFromEndpoint(context.TODO(), service, "127.0.0.1:3000", grpc.WithInsecure())

	if err :=e.Run(":3001");err != nil{
		fmt.Println(err)
	}
}
```