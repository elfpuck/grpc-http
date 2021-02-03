package elfHttp

import (
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
)

func recoveryMv() HandlerFunc {
	return func(c *Ctx) {
		defer func() {
			if err := recover(); err != nil {
				c.res = &resStruct{
					resData: nil,
					err:     status.Error(codes.Aborted, fmt.Sprintf("%s", err)),
				}
			}
		}()

		c.Next()
	}
}

func responseFormatMv() HandlerFunc {
	return func(c *Ctx) {

		c.Next()

		c.engine.ResponseFormat(nil)
		c.setResHeader()
		c.writer.Header().Set("Content-Type", "application/json")
		result := c.engine.responseFormatFunc(c, c.res.resData, c.res.err)
		encoder := json.NewEncoder(c.writer)
		if err := encoder.Encode(result); err != nil {
			http.Error(c.writer, err.Error(), 500)
		}
	}
}

func appendCtxHandlersMv() HandlerFunc {
	return func(c *Ctx) {
		if c.Request.Method != "POST" {
			c.Result(nil, status.Error(codes.InvalidArgument, "Request Should Use POST And application/json"))
			return
		}
		handlers, ok := c.engine.methodHandlersMap[c.routePath]
		if !ok {
			c.Result(nil, status.Error(codes.NotFound, "Request Error: "+c.routePath))
			return
		}
		c.handlers = append(c.handlers, handlers...)
		c.Context = metadata.NewIncomingContext(c.Context, c.ReqHeaderMD)

		c.Next()
	}
}
