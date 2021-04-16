package main

const TEMPLATE = `{
    "openapi":"3.0.0",
    "info":{
{{- range .Info }}
		"{{ .Name }}": {{ .Value|unescaped }}{{ .EndComma }}
{{- end}}
    },
    "servers":[
{{- range $k, $v := .Servers }}
		{{ .Value|unescaped }}{{ .EndComma }}
{{- end}}
    ],
    "security":[
{{- range .Security }}
		{ "{{ .Name }}" : [{{ .Value|unescaped }}] }{{ .EndComma }}
{{- end}}
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
		"securitySchemes":{
{{- range .SecuritySchemas }}
			"{{ .Name }}": {{ .Value|unescaped }}{{ .EndComma}}
{{- end}}
		},
      	"schemas": {
{{- range .Schemas }}
          	"{{ .Name }}": {{ .Value|unescaped }}{{ .EndComma}}
{{- end}}
		}
  	}
}`
