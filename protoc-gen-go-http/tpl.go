package main

const TEMPLATE = `
{{range .Services }}
// 注册 {{ .Name }}
func Register{{ .Name }}FromEndpoint(ctx context.Context, service *elfHttp.Service, endpoint string, opts ...grpc.DialOption) error {
	conn, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		return err
	}
	return register{{ .Name }}HandlerClient(service, New{{ .Name }}Client(conn))
}

// 注册Method
func register{{ .Name }}HandlerClient(s *elfHttp.Service, client {{ .Name }}Client) error {
{{range .Methods }}
	s.Method("{{ .Name }}", func(c *elfHttp.Ctx) {
        params := new({{ .Input}})
        if err := c.Unmarshal(params); err != nil{
			c.Result(nil, err)
			return
		}
		c.Result(client.{{ .Name }}(c.Context, params, grpc.Header(&c.ResHeaderMD), grpc.Trailer(&c.ResTrailerMD)))
	})
{{end}}
return nil
}
{{end}}
`
