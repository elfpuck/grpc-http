package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elfpuck/grpc-http/elfHttp"
	"github.com/elfpuck/grpc-http/protoc-gen-swagger/options"
	"google.golang.org/protobuf/compiler/protogen"
	"html/template"
	"sort"
	"strings"
)

func parseInfo(v interface{})( []*objStruct, error){
	data := make([]*objStruct, 0 , 20)
	dataMap := elfHttp.H{}
	jsonByte, _ := json.Marshal(v)
	json.Unmarshal(jsonByte, &dataMap)
	for k, v := range dataMap{
		dataByte, _ := json.Marshal(v)
		data = append(data, &objStruct{
			Name:     k,
			Value:    string(dataByte),
			EndComma: ",",
		})
	}

	return tidyObjStruct(data), nil
}

func parseServer(v interface{})([]*objStruct, error){
	data := make([]*objStruct, 0 , 20)
	var dataArr []interface{}
	jsonByte, _ := json.Marshal(v)
	json.Unmarshal(jsonByte, &dataArr)
	for _, v := range dataArr{
		dataByte, _ := json.Marshal(v)
		data = append(data, &objStruct{
			Value:    string(dataByte),
			EndComma: ",",
		})
	}

	return tidyObjStruct(data), nil
}

func parseSecurity(v interface{}) ([]*objStruct, error) {
	data := make([]*objStruct, 0 , 20)
	var dataArr []options.Security
	jsonByte, _ := json.Marshal(v)
	json.Unmarshal(jsonByte, &dataArr)
	for _, v := range dataArr{
		if v.Name == ""{
			panic("Security Lost name")
		}
		var valueData []string
		jsonByte, _ := json.Marshal(v.Scope)
		json.Unmarshal(jsonByte, &valueData)
		value := ""
		if len(valueData) != 0{
			jsonValue,_ := json.Marshal(strings.Join(valueData, ", "))
			value = string(jsonValue)
		}
		data = append(data, &objStruct{
			Name:     v.Name,
			Value:    value,
			EndComma: ",",
		})
	}

	return tidyObjStruct(data), nil
}

func parseSecuritySchema(v interface{})([]*objStruct, error){
	data := make([]*objStruct, 0 , 20)
	var dataArr []options.SecuritySchema
	jsonByte, _ := json.Marshal(v)
	json.Unmarshal(jsonByte, &dataArr)
	for _, v := range dataArr{
		if v.Name == "" || v.Type == nil{
			panic("Security Schema Lost name or type")
		}
		var typeData map[string]elfHttp.H
		jsonByte, _ := json.Marshal(v.Type)
		json.Unmarshal(jsonByte, &typeData)
		if len(typeData) != 1{
			panic("Security cant belong multi type" + fmt.Sprintf("%v", typeData))
		}
		for k2, v2 :=range typeData{
			v2["type"] = k2
			dataByte, _ := json.Marshal(v2)
			data = append(data, &objStruct{
				Name:     v.Name,
				Value:    string(dataByte),
				EndComma: ",",
			})
		}
	}
	return tidyObjStruct(data), nil
}

func parseFormatRes(v interface{}) (map[string]elfHttp.H, string) {

	res := map[string]elfHttp.H{}
	var dataKey string

	var dataArr []options.FormatRes
	jsonByte, _ := json.Marshal(v)
	json.Unmarshal(jsonByte, &dataArr)
	if dataArr == nil{
		res = map[string]elfHttp.H{
			"Data": {"type": "object"},
			"RetCode": {"type": "integer"},
			"Message": {"type": "string"},
		}
		dataKey = "Data"
		return res, dataKey
	}
	for _, v := range dataArr {
		if v.Key == "" || v.Type == "" {
			panic("formatRes Lost key or type")
		}
		if v.Primary == true{
			dataKey = v.Key
		}
		res[v.Key] = elfHttp.H{
			"type": v.Type,
		}
	}
	if dataKey == ""{
		panic("formatRes Lost primary true ")
	}
	return res, dataKey
}

func tidyObjStruct(data []*objStruct) []*objStruct {
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name > data[j].Name
	})
	if len(data) > 0 {
		data[len(data)-1].EndComma = ""
	}
	return data
}

func transMessageToSwagger(message *protogen.Message) elfHttp.H {
	baseSchema := schemaProperty{
		Description: currentStr(message.Comments.Leading.String(), message.Comments.Trailing.String()),
		Type:        "object",
		Properties:  map[string]schemaProperty{},
	}
	for _, v := range parseFields(message.Fields) {
		baseSchema.Properties[v.Name] = v.Property
	}
	dataByte, _ := json.Marshal(baseSchema)
	res := elfHttp.H{}
	json.Unmarshal(dataByte, &res)
	return res
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
