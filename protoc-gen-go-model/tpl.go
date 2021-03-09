package main

const TEMPLATE = `
// {{ .Name }}反射结构
type ReflectMessageRequest struct{}
type ReflectMessageResponse struct{}

func ReflectValueOfRequest () reflect.Value {
	return reflect.ValueOf(&ReflectMessageRequest{})
}

func ReflectValueOfResponse () reflect.Value {
	return reflect.ValueOf(&ReflectMessageResponse{})
}

{{- range .Methods }}
func (*ReflectMessageRequest) {{ .Name }}Request() proto.Message {
	return &{{ .Input }}{}
}
func (*ReflectMessageRequest) {{ .Name }}Response() proto.Message {
	return &{{ .Output }}{}
}
{{- end}}

{{- range .Methods }}
func (*ReflectMessageResponse) {{ .Name }}Response() proto.Message {
	return &{{ .Output }}{}
}
{{- end}}
`
