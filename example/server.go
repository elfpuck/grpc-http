package main

import (
	"github.com/elfpuck/grpc-http/example/controller"
	demo "github.com/elfpuck/grpc-http/example/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:3000")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	demo.RegisterDemoServer(s, &controller.Demo)
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		panic( err)
	}
}