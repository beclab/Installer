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
log_date=$(date +"%Y-%m-%d")
log_name="uninstall_${log_date}.log"
args=""
if [ "${os_type}" == "Darwin" ]; then
	args=" --minikube"
fi

sudo -E /bin/bash -c "terminus-cli terminus uninstall --base-dir $base_dir --phase $phase $args | tee -a $base_dir/logs/$log_name"

`)))
