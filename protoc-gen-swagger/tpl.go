package main

const TEMPLATE = `
openapi: 3.0.0
info:
  title: {{ .SourceName }}
  description: swagger generated by https://github.com/elfpuck/grpc-http/protoc-gen-swagger 
  version: 1.0.0
paths: 
{{- range .ApiArr }}
  {{ .RoutePath }}:
    post:
      tags:
        - {{ .Tag }}
      summary: {{ .Summary }}      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/{{ .SchemaReq }}'
      responses:
        '200':
          description: A JSON object
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/{{ .SchemaRes }}'
{{- end}}
components:
  schemas: 
{{- range .SchemaArr }}
    {{ .Name }}:
{{- if .Comments }}
      description: {{ .Comments }}
{{- end}}
      type: object
{{- if .Params }} 
      properties: 
{{- range .Params }}
        {{ .ParamsName }}:
{{- if .Comments }}
          description: {{ .Comments }}
{{- end}}
          type: {{ .ParamsType }}
{{- if .SonParams }}
          properties: 
{{- range .SonParams }}
            {{ .ParamsName }}:
{{- if .Comments }}
              description: {{ .Comments }}
{{- end}}
              type: {{ .ParamsType }}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
{{- end}}
`
