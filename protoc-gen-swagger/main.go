package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/imdario/mergo"
	"google.golang.org/protobuf/compiler/protogen"
	"html/template"
	"path"
	"regexp"
	"strings"
)

var (
	versionFlag = flag.Bool("version", false, "print the current version")
)

const (
	version           = "1.0.0"
	toolName          = "protoc-gen-swagger"
	schemaReqPrefix   = "Req__"
	schemaResPrefix   = "Res__"
	swaggerRegex      = "^//\\s*@swagger_(\\S*)\\s*(.*)$"
	swaggerFormatData = "{{ .Data }}"
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("%v %v\n", toolName, version)
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}
			generateFile(gen, f)
		}
		return nil
	})
}

func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + ".swagger.json"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)

	generateFileContent(file, g)
	return g
}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	data := PackageData{
		PathArr: make([]*pathStruct, 0, 20),
		PropMap: map[string]string{
			"info.title":   "\"" + file.Proto.GetPackage() + "\"",
			"info.version": "\"1.0.0\"",
		},
		SchemaMap: map[string]string{},
	}

	for _, service := range file.Services {

		addProp(service.Comments.Leading.String(), &data)

		for _, method := range service.Methods {

			api := pathStruct{
				Summary:   currentStr(method.Comments.Leading.String()),
				RoutePath: path.Join("/", fmt.Sprintf("%v", file.GoPackageName), method.GoName),
				SchemaReq: schemaReqPrefix + method.Input.GoIdent.GoName,
				SchemaRes: schemaResPrefix + method.Output.GoIdent.GoName,
				Tag:       service.GoName,
				EndComma:  ",",
			}

			data.PathArr = append(data.PathArr, &api)

			addSchema(api.SchemaReq, method.Input, &data)
			addSchema(api.SchemaRes, method.Output, &data)
		}
	}

	parseProp(&data)
	parseSchema(&data)

	if len(data.PathArr) > 0 {
		data.PathArr[len(data.PathArr)-1].EndComma = ""
	}

	g.P(executeTemplate(&data))
}

func addProp(comment string, data *PackageData) {
	propMap := map[string]string{}
	for _, s := range strings.Split(comment, "\n") {
		swaggerReg := regexp.MustCompile(swaggerRegex)
		baseSplit := swaggerReg.FindStringSubmatch(s)
		if len(baseSplit) < 2 {
			continue
		}
		propMap[baseSplit[1]] = baseSplit[2]
	}
	mergo.Merge(&data.PropMap, propMap, mergo.WithOverride)
}

func parseProp(data *PackageData) {
	for k, v := range data.PropMap {
		splitKey := strings.Split(k, ".")
		if len(splitKey) == 1 {
			if splitKey[0] != "openapi" && splitKey[0] != "info" && splitKey[0] != "paths" || splitKey[0] != "components" {
				data.PropArr = append(data.PropArr, &propStruct{
					Name:  splitKey[0],
					Value: v,
				})
			}
			continue
		}
		switch splitKey[0] {
		case "info":
			if splitKey[1] != "license" {
				data.InfoPropArr = append(data.InfoPropArr, &propStruct{
					Name:  splitKey[1],
					Value: v,
				})
			}
		case "components":
			if splitKey[1] != "schemas" {
				data.ComponentsPropArr = append(data.ComponentsPropArr, &propStruct{
					Name:  splitKey[1],
					Value: v,
				})
			}
		}
	}
}

func addSchema(key string, message *protogen.Message, data *PackageData) {
	baseSchema := schemaProperty{
		Description: currentStr(message.Comments.Leading.String(), message.Comments.Trailing.String()),
		Type:        "object",
		Properties:  map[string]schemaProperty{},
	}

	for _, v := range parseFields(message.Fields) {
		baseSchema.Properties[v.Name] = v.Property
	}

	dataByte, _ := json.Marshal(baseSchema)

	data.SchemaMap[key] = string(dataByte)
}

func parseSchema(data *PackageData) {
	formatReq, formatReqExists := data.PropMap["format.req"]
	formatRes, formatResExists := data.PropMap["format.res"]
	for k, v := range data.SchemaMap {
		if formatReqExists && strings.HasPrefix(k, schemaReqPrefix) {
			v = "{ \"type\": \"object\", \"properties\": " + strings.Replace(formatReq, swaggerFormatData, v, -1) + "}"
		}
		if formatResExists && strings.HasPrefix(k, schemaResPrefix) {
			v = "{ \"type\": \"object\", \"properties\": " + strings.Replace(formatRes, "{{ .Data }}", v, -1) + "}"
		}
		data.SchemaArr = append(data.SchemaArr, &schemaStruct{
			Name:     k,
			Value:    v,
			EndComma: ",",
		})
	}

	if len(data.SchemaArr) > 0 {
		data.SchemaArr[len(data.SchemaArr)-1].EndComma = ""
	}
}

func executeTemplate(data *PackageData) string {
	t := template.Must(template.New("swagger.json").Funcs(
		template.FuncMap{"unescaped": func(str string) template.HTML {
			return template.HTML(str)
		}}).Parse(TEMPLATE))
	res := new(bytes.Buffer)
	if err := t.Execute(res, data); err != nil {
		panic(err)
	}
	return res.String()
}

func currentStr(str ...string) string {
	for _, v := range str {
		if v != "" {
			return strings.Replace(v, "\n", "  ", -1)
		}
	}
	return ""
}

func switchType(str string) string {
	switch str {
	case "int32", "uint32", "int64", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		return "integer"
	case "double", "float":
		return "number"
	case "bytes":
		return "string"
	case "bool":
		return "boolean"
	default:
		return str
	}
	return str
}

func parseFields(fields []*protogen.Field) []*schemaParams {
	res := make([]*schemaParams, 0, len(fields))
	for _, v := range fields {
		res = append(res, &schemaParams{
			Name:     v.Desc.JSONName(),
			Property: parseField(v),
			EndComma: ",",
		})
	}
	if len(res) > 0 {
		res[len(res)-1].EndComma = ""
	}
	return res
}

func parseField(field *protogen.Field) schemaProperty {
	data := schemaProperty{
		Description: currentStr(field.Comments.Leading.String(), field.Comments.Trailing.String()),
	}
	if field.Desc.IsMap() {
		data.Type = "object"
	} else if field.Desc.IsList() {
		data.Type = "array"
		data.Items = &schemaProperty{}
		sonFieldType(field, data.Items)
	} else {
		sonFieldType(field, &data)
	}
	return data
}

func sonFieldType(field *protogen.Field, schema *schemaProperty) {

	fieldType := switchType(field.Desc.Kind().String())
	if fieldType == "message" {
		schema.Type = "object"
		schema.Properties = sonObjectFieldType(field.Message.Fields)
		return
	}
	if fieldType == "enum" {
		schema.Type = "string"
		schema.Enum = sonEnumFieldType(field)
		return
	}
	schema.Type = fieldType
}

func sonEnumFieldType(field *protogen.Field) []string {
	res := make([]string, 0, len(field.Enum.Values))
	for _, v := range field.Enum.Values {
		res = append(res, strings.TrimPrefix(v.GoIdent.GoName, field.Enum.GoIdent.GoName+"_"))
	}
	return res
}

func sonObjectFieldType(fields []*protogen.Field) map[string]schemaProperty {
	res := map[string]schemaProperty{}
	for _, v := range fields {
		res[v.Desc.JSONName()] = parseField(v)
	}
	return res
}

type PackageData struct {
	PropMap           map[string]string
	PathArr           []*pathStruct
	PropArr           []*propStruct
	InfoPropArr       []*propStruct
	ComponentsPropArr []*propStruct
	SecurityPropArr   []*propStruct
	SchemaMap         map[string]string
	SchemaArr         []*schemaStruct
}

type propStruct struct {
	Name  string
	Value string
}

type pathStruct struct {
	Summary   string
	RoutePath string
	SchemaReq string
	SchemaRes string
	Tag       string
	EndComma  string
}

type schemaStruct struct {
	Name     string
	Value    string
	EndComma string
}

type schemaParams struct {
	Name     string
	Property schemaProperty
	EndComma string
}

type schemaProperty struct {
	Type        string                    `json:"type,omitempty"`
	Description string                    `json:"description,omitempty"`
	Properties  map[string]schemaProperty `json:"properties,omitempty"`
	Items       *schemaProperty           `json:"items,omitempty"`
	Enum        []string                  `json:"enum,omitempty"`
}
