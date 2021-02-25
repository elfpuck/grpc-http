## protoc-gen-go-http

## Installation
```shell script
go get -u github.com/golang/protobuf/protoc-gen-go            #生成go代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-go-http     #生成grpc-http 代码
go get -u github.com/elfpuck/grpc-http/protoc-gen-swagger     #生成swagger
```

## Generate
```shell script
protoc -I=. --go_out=plugins=grpc:. --http_out=. --swagger_out=. xxx.proto
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

### Swagger 生成
```proto
/**
  @swagger_servers [{"url":"https://www.baidu.com","description":"请求服务地址"}]
  @swagger_security [{"tokenAuth":[]}]
  @swagger_info.version "1.0.1"
  @swagger_components.securitySchemes { "tokenAuth": {"type": "apiKey", "in": "cookie", "name":"token"}}
  @swagger_format.res {"Data": {{ .Data }}, "Code": {"type": "integer"}, "Message": {"type": "string"}
 */
service Demo {}
```
* `@swagger_servers` 请求服务地址

* `@swagger_security` 请求全局安全策略

* `@swagger_info.version` 服务版本 | 默认服务版本为swagger 生成时间

* `@swagger_components.securitySchemes` 安全策略

* `@swagger_format.req` 修改网关请求内容  {{ .Data }} 为`pb`请求内容

* `@swagger_format.res` 修改网关返回内容  {{ .Data }} 为`pb`返回内容