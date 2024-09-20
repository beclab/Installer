package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var TerminusUninstallScriptValues = template.Must(template.New("terminus-uninstall.sh").Parse(
	dedent.Dedent(`#!/bin/sh
set -x

os_type=$(uname -s)
base_dir={{ .BaseDir }}
phase={{ .Phase }}
installer_path="v{{ .Version }}"
args=""
if [ "${os_type}" == "Darwin" ]; then
	args=" --minikube"
fi

sudo -E /bin/bash -c "terminus-cli terminus uninstall --version {{ .Version }} --base-dir $base_dir --phase $phase $args | tee $base_dir/versions/$installer_path/logs/uninstall.log"

`)))
