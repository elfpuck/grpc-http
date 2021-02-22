package elfHttp

import (
	"fmt"
	"google.golang.org/grpc"
)

type Service struct {
	name              string
	dialOptions       []grpc.DialOption
	handlers          handlersChain
	methodHandlersMap map[string]handlersChain
	engine            *Engine
}

func (g *Service) UseGrpc(opts ...grpc.DialOption)  {
	if len(opts) == 0 {
		return
	}
	if g.dialOptions == nil{
		g.dialOptions = []grpc.DialOption{}
	}
	g.dialOptions = append(g.dialOptions, opts...)
}

func (g *Service) getDialOptions(opts ...grpc.DialOption) []grpc.DialOption {
	var dialArr []grpc.DialOption
	dialArr = append(dialArr, g.engine.globalGroup.dialOptions...)
	dialArr = append(dialArr, g.dialOptions...)
	dialArr = append(dialArr, opts...)
	return dialArr
}

func (g *Service) Use(middleware ...HandlerFunc) {
	if len(middleware) == 0 {
		return
	}
	if g.handlers == nil {
		g.handlers = make(handlersChain, 0, len(middleware))
	}
	g.handlers = append(g.handlers, middleware...)
}

func (g *Service) Method(method string, handlers ...HandlerFunc) {
	if _, ok := g.methodHandlersMap[method]; ok {
		panic(fmt.Sprintf("method conflict: %s", method))
	}
	g.methodHandlersMap[method] = handlers
}

func (g *Service) Dial(endpoint string, opts ...grpc.DialOption)  (conn *grpc.ClientConn, err error)  {
	return grpc.DialContext(g.engine.Context, endpoint, g.getDialOptions(opts...)...)
}