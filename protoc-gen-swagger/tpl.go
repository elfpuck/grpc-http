package main

const TEMPLATE = `{
    "openapi":"3.0.0",
    "info":{
{{- range .InfoPropArr }}
		"{{ .Name }}": {{ .Value|unescaped }},
{{- end}}
        "license": {}
    },
{{- range .PropArr }}
	"{{ .Name }}": {{ .Value|unescaped }},
{{- end}}
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
{{- range .ComponentsPropArr }}
		"{{ .Name }}": {{ .Value|unescaped }},
{{- end}}
      	"schemas": {
{{- range .SchemaArr }}
          	"{{ .Name }}": {{ .Value|unescaped }}{{ .EndComma}}
{{- end}}
		}
  	}
}`
