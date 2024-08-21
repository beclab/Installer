package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var SettingsValue = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`namespace:
  name: 'user-space-{{ .UserName }}'
  role: admin

{{ if .Storage }}
{{ range $key, $value := .Storage }}
{{ $key }}: {{ $value }}
{{ end }}
{{ end }}

user:
  name: '{{ .UserName }}'
`),
))
