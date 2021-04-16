## protoc-gen-go-http
> 本项目部分参考`grpc-ecosystem/grpc-gateway`,向该项目致敬。 

## Installation
```shell script
go get -u github.com/golang/protobuf/protoc-gen-go            #生成go代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-go-http     #生成grpc-http 代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-go-model    #生成grpc-model 代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-swagger     #生成swagger
```

## Generate
```shell script
protoc -I=. -I=${GOPATH}/src/github.com/elfpuck/grpc-http --go_out=plugins=grpc:. --go-http_out=. --go-model_out=. --swagger_out=. xxx.proto
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
	demo.RegisterDemoFromEndpoint( service, "127.0.0.1:3000")

	if err :=e.Run(":3001");err != nil{
		fmt.Println(err)
	}
}
```

### Swagger
```proto
syntax = "proto3";

package test;

option go_package = "./pb";

import "protoc-gen-swagger/options/swagger.proto";

option (elfpuck.options.swagger) = {
  formatRes: [
    {
      key: "RetCode",
      type: "string"
    },
    {
      key: "Message",
      type: "string"
    },
    {
      key: "Data",
      type: "object",
      primary: true
    }
  ]
  info: {
    title: "grpc-gateway",
    version: "1.0",
    description: "多grpc gateway平台",
    contact:{
      name: "flynn",
      email: "shanquan54@gmail.com"
    },
    license:{
      name: "MIT"
    },
  },
  security: [
    {
      name: 'tokenAuth'
      scope: [],
    }
  ]
  servers: [
    {
      url: "http://127.0.0.1:3000",
      description: "本地测试",
    },
    {
      url: "https://www.abc.cn",
      description: "线上环境",
    }
  ],
  securitySchemes: [
    {
      name: "tokenAuth"
      type: {
        apiKey: {
          in: "cookie",
          name: "token"
        }
      }
    }
  ]
};

service Test{}
```
