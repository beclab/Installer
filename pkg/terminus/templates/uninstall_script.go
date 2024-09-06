package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var TerminusUninstallScriptValues = template.Must(template.New("terminus-uninstall.sh").Parse(
	dedent.Dedent(`#!/bin/sh
set -x

sudo -E /bin/bash -c "terminus-cli terminus uninstall --base-dir {{ .BaseDir }} --phase {{ .Phase }}"

`)))
