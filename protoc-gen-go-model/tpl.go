package main

const TEMPLATE = `
// {{ .Name }}反射结构
type ReflectMessageRequest struct{}
type ReflectMessageResponse struct{}

func ReflectValueOf () reflect.Value {
	return reflect.ValueOf(&ReflectMessageRequest{})
}

{{- range .Methods }}
func (*ReflectMessageRequest) {{ .Name }}() proto.Message {
	return &{{ .Input }}{}
}
{{- end}}

{{- range .Methods }}
func (*ReflectMessageResponse) {{ .Name }}() proto.Message {
	return &{{ .Output }}{}
}
{{- end}}
`
