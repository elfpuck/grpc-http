package controller

import (
	"context"
	demo "github.com/elfpuck/grpc-http/example/pb"
)

type Demo struct {}

func (d *Demo)Echo(ctx context.Context,  ping *demo.Ping) () {

}
