package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var AppValues = template.Must(template.New("values.yaml").Parse(
	dedent.Dedent(``),
))
