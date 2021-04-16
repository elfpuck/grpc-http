package elfHttp

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/textproto"
	"strings"
)

const abortIndex int8 = math.MaxInt8 / 2

type Ctx struct {
	routePath       string
	index           int8
	bodyByte        []byte
	handlers        handlersChain
	Context         context.Context
	writer          http.ResponseWriter
	Request         *http.Request
	headerMDPrefix  string
	trailerMDPrefix string
	ReqHeaderMD     metadata.MD
	ResHeaderMD     metadata.MD
	ResTrailerMD    metadata.MD
	res             *resStruct
	engine          *Engine
}

type resStruct struct {
	resData interface{}
	err     error
}

func (c *Ctx) reset() {
	c.Context = c.engine.Context
	c.index = -1
	c.writer = nil
	c.Request = nil
	c.bodyByte = nil
	c.handlers = c.handlers[0:len(c.engine.globalGroup.handlers)]
	c.ReqHeaderMD = nil
	c.ResHeaderMD = metadata.MD{}
	c.ResTrailerMD = metadata.MD{}
	c.res = nil
}

func (c *Ctx) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Ctx) RoutePath() string {
	return c.routePath
}

// 添加header信息到md上
func (c *Ctx) annotateContext() {
	var pairs []string
	for key, vals := range c.Request.Header {
		key = textproto.CanonicalMIMEHeaderKey(key)
		if strings.HasPrefix(key, c.headerMDPrefix) {
			for _, val := range vals {
				pairs = append(pairs, key[len(c.headerMDPrefix):], val)
			}
		}
	}
	c.ReqHeaderMD = metadata.Pairs(pairs...)
}

func (c *Ctx) setResHeader() {
	for key, vals := range c.ResHeaderMD {
		key = textproto.CanonicalMIMEHeaderKey(key)
		if strings.HasPrefix(key, c.headerMDPrefix) {
			for _, val := range vals {
				c.writer.Header().Add(key[len(c.headerMDPrefix):], val)
			}
		}
	}

	for k := range c.ResTrailerMD {
		tKey := textproto.CanonicalMIMEHeaderKey(c.trailerMDPrefix + k)
		c.writer.Header().Add("Trailer", tKey)
	}
}

// get BodyByte, please using SetBodyByte after GetBodyByte()
func (c *Ctx) BodyByte(f func(oldBodyByte []byte) (newBodyByte []byte)) []byte {
	if f == nil {
		if c.bodyByte != nil {
			return c.bodyByte
		}
		reqBodyByte, _ := ioutil.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyByte))
		c.bodyByte = reqBodyByte
		return c.bodyByte
	}

	oldBodyByte := c.bodyByte
	if oldBodyByte == nil {
		reqBodyByte, _ := ioutil.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()
		oldBodyByte = reqBodyByte
	}
	res := f(oldBodyByte)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(res))
	c.bodyByte = res
	return c.bodyByte
}

func (c *Ctx) Unmarshal(req interface{}) error {
	if err := json.Unmarshal(c.BodyByte(nil), &req); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}

// 设置返回值
func (c *Ctx) Result(resData interface{}, err error) {
	c.index = abortIndex

	if err == nil && resData != nil && c.engine.jsonPb != nil {
		switch m := resData.(type) {
		case proto.Message:
			str, marShalErr := c.engine.jsonPb.MarshalToString(m)
			if err != nil {
				resData = nil
				err = marShalErr
			}
			res := H{}
			json.Unmarshal([]byte(str), &res)
			resData = res
		}
	}

	c.res = &resStruct{
		resData: resData,
		err:     err,
	}
}

func (c *Ctx) ClientIP() string {
	clientIP := ""
	if clientIP = c.Request.Header.Get("X-Forwarded-For"); clientIP == "" {
		if clientIP = strings.TrimSpace(c.Request.Header.Get("X-Real-Ip")); clientIP == "" {
			if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
				clientIP = ip
			}
		}
	}
	return clientIP
}
