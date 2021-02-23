package main

const TEMPLATE = `{
    "openapi":"3.0.0",
    "version":"1.0.0",
    "info":{
        "version":"1.0.0",
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
{{- range .SchemaArr }}
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
	}
  }
}`
