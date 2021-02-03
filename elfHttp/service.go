package elfHttp

import (
	"fmt"
)

type Service struct {
	name              string
	handlers          handlersChain
	methodHandlersMap map[string]handlersChain
	engine            *Engine
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
