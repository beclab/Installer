package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var WSLConfigValue = template.Must(template.New(".wslconfig").Parse(
	dedent.Dedent(`[wsl2]
memory=16GB
swap=0GB
`),
))
