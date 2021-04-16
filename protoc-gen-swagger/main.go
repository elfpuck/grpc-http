package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/elfpuck/grpc-http/elfHttp"
	"github.com/elfpuck/grpc-http/protoc-gen-swagger/options"
	"github.com/golang/protobuf/proto"
	"google.golang.org/protobuf/compiler/protogen"
	"path"
	"strings"
)

var (
	versionFlag = flag.Bool("version", false, "print the current version")
)

const (
	version           = "1.0.0"
	toolName          = "protoc-gen-swagger"
	schemaReqPrefix   = "req__"
	schemaResPrefix   = "res__"
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
		Info: make([]*objStruct, 0 , 20),
	}

	ext, err := proto.GetExtension(file.Proto.Options, options.E_Swagger)
	var opts *options.Swagger
	if err != nil{
		opts = &options.Swagger{}
	}else {
		tempOpts, ok := ext.(*options.Swagger)
		if !ok{
			panic("swagger Swagger 错误" + fmt.Sprintf("%v", opts))
		}
		opts = tempOpts
	}

	if opts.Info == nil{
		opts.Info = &options.Info{
			Title: file.Proto.GetPackage(),
		}
	}

	data.Info, _ = parseInfo(opts.Info)
	data.Servers,_ = parseServer(opts.Servers)
	data.SecuritySchemas, _ = parseSecuritySchema(opts.SecuritySchemes)
	data.Security, _ = parseSecurity(opts.Security)
	resMap, resKey := parseFormatRes(opts.FormatRes)

	for _, service := range file.Services {
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

			reqBodyByte, _ := json.Marshal(transMessageToSwagger(method.Input))

			resMap[resKey] = transMessageToSwagger(method.Output)
			resBodyByte, _ := json.Marshal(elfHttp.H{"type": "object","properties" : resMap})

			data.Schemas = append(data.Schemas, &objStruct{
				Name:     api.SchemaReq,
				Value:    string(reqBodyByte),
				EndComma: ",",
			}, &objStruct{
				Name:     api.SchemaRes,
				Value:    string(resBodyByte),
				EndComma: ",",
			})
		}
	}

	data.Schemas = tidyObjStruct(data.Schemas)

	if len(data.PathArr) > 0 {
		data.PathArr[len(data.PathArr)-1].EndComma = ""
	}

	g.P(executeTemplate(&data))
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
	Info       	      []*objStruct
	Servers			  []*objStruct
	Security          []*objStruct
	Schemas			  []*objStruct
	SecuritySchemas   []*objStruct
	PathArr           []*pathStruct
}

type pathStruct struct {
	Summary   string
	RoutePath string
	SchemaReq string
	SchemaRes string
	Tag       string
	EndComma  string
}

type objStruct struct {
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
