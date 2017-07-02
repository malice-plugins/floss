package main

const tpl = `#### Floss
{{- if .Results.ASCIIStrings}}

##### ASCII Strings
{{range .Results.ASCIIStrings}}
  - {{ . -}}
{{end}}{{end-}}
{{- if .Results.UTF16Strings}}

##### UTF-16 Strings
{{range .Results.UTF16Strings}}
  - {{ . -}}
{{end}}{{end-}}
{{- if .Results.DecodedStrings}}

##### Decoded Strings
{{range .Results.DecodedStrings}}
  Location: {{ .Location -}}
  {{range .Strings}}
    - {{ . -}}
  {{end}}
{{end}}{{end-}}
{{- if .Results.StackStrings}}

##### Stack Strings
{{range .Results.StackStrings}}
  - {{ . -}}
{{end}}{{end}}
`
