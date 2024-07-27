package templates

import (
	"html/template"

	"github.com/lithammer/dedent"
)

var SettingsValue = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`namespace:
  name: 'user-space-{{ .UserName }}'
  role: admin

cluster_id: {{ .ClusterId }}
s3_sts: {{ .StorageToken }}
s3_ak: {{ .StorageAccessKey }}
s3_sk: {{ .StorageSecretKey }}

user:
  name: '{{ .UserName }}'
	`),
))
