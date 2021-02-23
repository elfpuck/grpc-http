package elfHttp

import (
	"context"
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"path"
	"sync"
)

type H map[string]interface{}
type HandlerFunc func(ctx *Ctx)
type handlersChain []HandlerFunc
type responseFormatFunc func(c *Ctx, data interface{}, err error) H

type Engine struct {
	Context            context.Context
	headerMDPrefix     string
	trailerMDPrefix    string
	globalGroup        *Service
	groupList          []*Service
	methodHandlersMap  map[string]handlersChain
	responseFormatFunc responseFormatFunc
	pool               sync.Pool
}

func New() *Engine {
	engine := &Engine{
		Context:           context.TODO(),
		headerMDPrefix:    "Grpc-Metadata-",
		trailerMDPrefix:   "Grpc-Trailer-",
		groupList:         make([]*Service, 0, 2),
		methodHandlersMap: map[string]handlersChain{},
	}
	engine.globalGroup = &Service{}

	// 将res写入writer
	engine.Use(responseFormatMv(), loggerMv(), recoveryMv())
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() *Ctx {
	c := &Ctx{
		index:        -1,
		Context:      e.Context,
		ResHeaderMD:  metadata.MD{},
		ResTrailerMD: metadata.MD{},
		handlers:     append(handlersChain{}, e.globalGroup.handlers...),
		engine:       e,
	}
	return c
}

func (e *Engine) ChangeRoute(f func(ctx *Ctx) string) {
	if f != nil {
		e.Use(func(c *Ctx) {
			routePath := f(c)
			if routePath != "" {
				c.routePath = routePath
			}
		})
	}
}

// format result body
func (e *Engine) ResponseFormat(f responseFormatFunc) {
	if f != nil && e.responseFormatFunc == nil {
		e.responseFormatFunc = f
		return
	}
	if e.responseFormatFunc == nil {
		e.responseFormatFunc = e.defaultResponseFormat
	}
}

func (e *Engine) defaultResponseFormat(c *Ctx, res interface{}, err error) H {
	result := H{}
	if err != nil {
		status, ok := status.FromError(err)
		if ok {
			result["Code"] = status.Code()
			result["Message"] = status.Message()
		} else {
			result["Code"] = codes.Internal
			result["Message"] = err.Error()
		}
	} else {
		result["Code"] = codes.OK
		result["Message"] = "ok"
		result["Data"] = res
	}
	return result
}

func (e *Engine) HeaderMDPrefix(f func(oldHeaderMDPrefix string) string) string {
	if f != nil {
		e.headerMDPrefix = f(e.headerMDPrefix)
	}
	return e.headerMDPrefix
}

func (e *Engine) TrailerMDPrefix(f func(oldTrailerMDPrefix string) string) string {
	if f != nil {
		e.trailerMDPrefix = f(e.trailerMDPrefix)
	}
	return e.trailerMDPrefix
}

func (e *Engine) Use(middleware ...HandlerFunc) {
	e.globalGroup.Use(middleware...)
}

func (e *Engine) UseGrpc(opts ...grpc.DialOption) {
	e.globalGroup.UseGrpc(opts...)
}

func (e *Engine) Service(serviceName string, handlers ...HandlerFunc) *Service {
	g := &Service{
		name:              serviceName,
		handlers:          handlers,
		methodHandlersMap: map[string]handlersChain{},
		engine:            e,
	}
	e.groupList = append(e.groupList, g)
	return g
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Ctx)

	c.headerMDPrefix = e.headerMDPrefix
	c.trailerMDPrefix = e.trailerMDPrefix
	c.routePath = req.URL.Path
	c.writer = w
	c.Request = req
	c.annotateContext()
	c.Next()
	c.reset()

	e.pool.Put(c)
}

func (e *Engine) mapApiGroupList() error {
	for _, group := range e.groupList {
		debugPrint("\nRegister Service %s\n", aurora.Green(group.name))
		for method, handlers := range group.methodHandlersMap {
			routePath := path.Join("/", group.name, method)
			debugPrint("Register Method %s\n", aurora.Cyan(routePath))
			if _, ok := e.methodHandlersMap[routePath]; ok {
				return errors.New(fmt.Sprintf("Service %s method conflict %s", group.name, method))
			}
			hdls := make(handlersChain, 0, len(handlers)+len(group.handlers))
			hdls = append(hdls, group.handlers...)
			hdls = append(hdls, handlers...)
			e.methodHandlersMap[routePath] = hdls
		}
		group = nil
	}
	e.groupList = nil
	return nil
}

func (e *Engine) Run(addr string) error {
	if err := e.mapApiGroupList(); err != nil {
		return err
	}
	// append action handlers
	e.Use(appendCtxHandlersMv())

	debugPrint("\nListening and serving HTTP on %s \n", aurora.Green(addr))
	return http.ListenAndServe(addr, e)
}

func (e *Engine) RunTLS(addr string, certFile, keyFile string) error {
	if err := e.mapApiGroupList(); err != nil {
		return err
	}
	// append action handlers
	e.Use(appendCtxHandlersMv())

	debugPrint("\nListening and serving HTTPS on %s \n", aurora.Green(addr))
	return http.ListenAndServeTLS(addr, certFile, keyFile,  e)
}