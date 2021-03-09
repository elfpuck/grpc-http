package main

const TEMPLATE = `
// {{ .Name }}反射结构
type reflectMessage struct{}

func ReflectValueOf () reflect.Value {
	return reflect.ValueOf(&reflectMessage{})
}

{{- range .Methods }}
func (*reflectMessage) {{ .Name }}() proto.Message {
	return &{{ .Input }}{}
}
{{- end}}
`
