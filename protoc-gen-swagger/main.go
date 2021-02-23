package main

import (
	"bytes"
	"flag"
	"fmt"
	"google.golang.org/protobuf/compiler/protogen"
	"html/template"
	"path"
	"strings"
	"time"
)

var (
	versionFlag = flag.Bool("version", false, "print the current version")
)

const (
	version  = "1.0.0"
	toolName = "protoc-gen-swagger"
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
		Title:       file.Proto.GetName(),
		Description: fmt.Sprintf("%s created by https://github.com/elfpuck/grpc-http/protoc-gen-swagger on %s", file.Proto.GetName(), time.Now().Format("2006-01-02 15:04:05")),
		PathArr:     make([]*pathStruct, 0, 20),
		SchemaArr:   make([]*schemaStruct, 0, 40),
	}

	for _, service := range file.Services {
		for _, method := range service.Methods {
			schemaReq := schemaStruct{
				Comments: currentStr(method.Input.Comments.Leading.String(), method.Input.Comments.Trailing.String()),
				Name:     fmt.Sprintf("%v%s", file.GoPackageName, method.Input.GoIdent.GoName),
				EndComma: ",",
			}
			schemaRes := schemaStruct{
				Comments: currentStr(method.Output.Comments.Leading.String(), method.Output.Comments.Trailing.String()),
				Name:     fmt.Sprintf("%v%s", file.GoPackageName, method.Output.GoIdent.GoName),
				EndComma: ",",
				Params: []*schemaParams{
					{
						Name: "Code",
						Property: "{\n\t\t\t\t\t\"type\": \"integer\",\n\t\t\t\t\t\"description\": \"返回状态码\"\n\t\t\t\t}",
						EndComma: ",",
					},
					{
						Name: "Message",
						Property: "{\n\t\t\t\t\t\"type\": \"string\",\n\t\t\t\t\t\"description\": \"报错内容\"\n\t\t\t\t}",
						EndComma: ",",
					},
				},
			}
			api := pathStruct{
				Summary:   currentStr(method.Comments.Leading.String()),
				RoutePath: path.Join("/", fmt.Sprintf("%v", file.GoPackageName), method.GoName),
				SchemaReq: schemaReq.Name,
				SchemaRes: schemaRes.Name,
				Tag:       service.GoName,
				EndComma: ",",
			}
			// 请求参数
			for _, v := range method.Input.Fields {
				//reqSchemaParamsStruct := schemaParamsStruct{
				//	Comments:   currentStr(v.Comments.Leading.String(), v.Comments.Trailing.String()),
				//	ParamsName: v.Desc.JSONName(),
				//	ParamsType: switchType(v.Desc.Kind().String()),
				//}

				schemaReq.Params = append(schemaReq.Params, &schemaParams{
					Name:     v.Desc.JSONName(),
					Property: "{}",
					EndComma: ",",
				})
			}

			if len(schemaReq.Params) > 0{
				schemaReq.Params[len(schemaReq.Params) - 1].EndComma = ""
			}

			//返回参数
			resData := schemaParams{
				Name: "Data",
				Property: "{}",
				EndComma: "",
			}
			//for _, v := range method.Output.Fields {
			//	sonSchema := schemaParamsStruct{
			//		Comments:   currentStr(v.Comments.Leading.String(), v.Comments.Trailing.String()),
			//		ParamsName: v.Desc.JSONName(),
			//		ParamsType: switchType(v.Desc.Kind().String()),
			//	}
			//	resSchemaParamsStruct.SonParams = append(resSchemaParamsStruct.SonParams, &sonSchema)
			//}

			schemaRes.Params = append(schemaRes.Params, &resData)

			data.SchemaArr = append(data.SchemaArr, &schemaReq, &schemaRes)
			data.PathArr = append(data.PathArr, &api)
		}
	}

	if len(data.PathArr) > 0{
		data.PathArr[len(data.PathArr) -1].EndComma = ""
	}
	if len(data.SchemaArr) > 0{
		data.SchemaArr[len(data.SchemaArr) -1].EndComma = ""
	}
	g.P(executeTemplate(&data))
}

func executeTemplate(data *PackageData) string {
	t := template.Must(template.New("swagger.json").Funcs(
		template.FuncMap{"unescaped": func(str string) template.HTML {
			return template.HTML(str)}}).Parse(TEMPLATE))
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
	case "message":
		return "object"
	case "int32", "uint32", "int64", "uint64", "sint32", "sint64", "fixed32", "fixed64", "sfixed32", "sfixed64":
		return "integer"
	case "double", "float":
		return "number"
	case "bytes":
		return "string"
	default:
		return str
	}
	return str
}

type PackageData struct {
	Title       string
	Description string
	PathArr     []*pathStruct
	SchemaArr   []*schemaStruct
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
	Comments string
	Name     string
	Params   []*schemaParams
	EndComma  string
}

type schemaParams struct {
	Name string
	Property string
	EndComma  string
}

type schemaProperty struct {
	Type string								`json:"type,omitempty"`
	Description string  					`json:"description,omitempty"`
	Properties map[string]schemaProperty	`json:"properties,omitempty"`
}
