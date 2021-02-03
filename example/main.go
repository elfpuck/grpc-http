package main

import (
	"context"
	"fmt"
	"github.com/elfpuck/grpc-http/elfHttp"
	demo "github.com/elfpuck/grpc-http/example/pb"
	"google.golang.org/grpc"
)

func main() {
	e := elfHttp.New()
	service := e.Service("demo")
	demo.RegisterDemoFromEndpoint(context.TODO(), service, "127.0.0.1:3000", grpc.WithInsecure())

	if err := e.Run(":3001"); err != nil {
		fmt.Println(err)
	}
}
