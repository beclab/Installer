package templates

import (
	"html/template"

	"github.com/lithammer/dedent"
)

var LauncherValues = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(`bfl:
  nodeport: 30883
  nodeport_ingress_http: 30083
  nodeport_ingress_https: 30082
  username: '{{ .UserName }}'
  admin_user: true
	`),
))
