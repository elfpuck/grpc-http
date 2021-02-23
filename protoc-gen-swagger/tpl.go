package main

const TEMPLATE = `{
    "openapi":"3.0.0",
    "version":"1.0.0",
    "info":{
        "version": "{{ .Version }}",
        "title":"{{ .Title }}",
        "description":"{{ .Description }}"
    },
    "servers":[

    ],
  "paths": { {{- range .PathArr }}
        "{{ .RoutePath }}":{
            "post":{
                "tags":[
                    "{{ .Tag }}"
                ],
                "summary":"{{ .Summary }}",
                "requestBody":{
                    "required":true,
                    "content":{
                        "application/json":{
                            "schema":{
                                "$ref":"#/components/schemas/{{ .SchemaReq }}"
                            }
                        }
                    }
                },
                "responses":{
                    "200":{
                        "description":"A JSON Object",
                        "content":{
                            "application/json":{
                                "schema":{
                                    "$ref":"#/components/schemas/{{ .SchemaRes }}"
                                }
                            }
                        }
                    }
                }
            }
        }{{ .EndComma }}
{{- end}}
  },
  "components": {
      "schemas": {
{{- range .SchemaReqArr }}
          	"{{ .Name }}":{
{{- if .Comments }}
      		"description": "{{ .Comments }}",
{{- end}}
			"type": "object",
            "properties": {
{{- range .Params }}
            	"{{ .Name }}": {{ .Property|unescaped }}{{ .EndComma }}
{{- end}}
			}
		  }{{ .EndComma }}
{{- end}}
{{- range .SchemaResArr }}
			"{{ .Name }}":{
				"type": "object",
            	"properties": {
					"Code": {
						"type": "integer",
						"description": "返回状态码"
					},
					"Message": {
						"type": "string",
						"description": "报错内容"
					},
					"Data": {
						"type": "object",
						"properties": {
{{- range .Params }}
            				"{{ .Name }}": {{ .Property|unescaped }}{{ .EndComma }}
{{- end}}
						}
					}
			}
		  }{{ .EndComma }}
{{- end}}
	}
  }
}`
