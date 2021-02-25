package elfHttp

import (
	"encoding/json"
	"fmt"
	"github.com/logrusorgru/aurora"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"time"
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

func loggerMv() HandlerFunc {
	return func(c *Ctx) {
		startTime := time.Now()
		c.Next()
		if IsDebugging() {
			endTime := time.Now()
			if c.res.err != nil {
				debugPrint("%s %20s %16s %5d ms | %s", time.Now().Format("2006-01-02 15:04:05"), aurora.Red(c.routePath), c.ClientIP(), endTime.Sub(startTime).Milliseconds(), aurora.Red(c.res.err.Error()))
			} else {
				debugPrint("%s %20s %16s %5d ms", time.Now().Format("2006-01-02 15:04:05"), aurora.Green(c.routePath), c.ClientIP(), endTime.Sub(startTime).Milliseconds())
			}
		}
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
		if c.Request.Method != "POST" || c.Request.Header.Get("content-type") != "application/json" {
			c.Result(nil, status.Error(codes.InvalidArgument, "Request Should Use POST And application/json"))
			return
		}
		handlers, ok := c.engine.methodHandlersMap[c.routePath]
		if !ok {
			c.Result(nil, status.Error(codes.NotFound, "Request Error: "+c.routePath))
			return
		}
		c.handlers = append(c.handlers, handlers...)
		c.Context = metadata.NewOutgoingContext(c.Context, c.ReqHeaderMD)

		c.Next()
	}
}
