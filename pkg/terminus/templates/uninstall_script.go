package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

var OlaresUninstallScriptValues = template.Must(template.New("olares-uninstall.sh").Parse(
	dedent.Dedent(`#!/bin/bash
set +x
os_type=$(uname -s)
base_dir={{ .BaseDir }}
version="{{ .Version }}"

args=""
phase="$1"
if [ -z "$phase" ]; then
  args+=" --all"
fi

if [[ ! -z "$phase" && x"$phase" != x"prepare" ]]; then
  echo "The parameter is incorrect, the parameter value is: prepare."
  exit 1
else
  args+=" --phase install"
fi

sudo -E /bin/bash -c "olares-cli terminus uninstall --version $version --base-dir $base_dir $args"

`)))
