package elfRpc

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"sync"
)

type handler interface {
	Register(s *grpc.Server)
}

type engine struct {
	sync.Once
	server *grpc.Server
	mv     []grpc.ServerOption
}

func New() *engine {
	engine := &engine{}
	return engine
}

func (e *engine) Use(opt ...grpc.ServerOption) {
	if e.server != nil {
		panic("cant Use this method after register")
	}
	e.mv = append(e.mv, opt...)
}

func (e *engine) Server() *grpc.Server {
	e.lazyInit()
	return e.server
}

func (e *engine) Run(network string, addr string) {
	e.lazyInit()
	lis, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	reflection.Register(e.server)
	fmt.Printf("\n\nListening RPC on %s %s\n", network, addr)
	if err := e.server.Serve(lis); err != nil {
		panic(err)
	}
}

func (e *engine) lazyInit() {
	e.Once.Do(func() {
		s := grpc.NewServer(e.mv...)
		e.server = s
	})
}
