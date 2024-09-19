package templates

import (
	"text/template"

	"github.com/lithammer/dedent"
)

// TerminusdEnv defines the template of terminusd's env.
var TerminusdEnv = template.Must(template.New("terminusd.service.env").Parse(
	dedent.Dedent(`# Environment file for terminusd
INSTALLED_VERSION={{ .Version }}
KUBE_TYPE={{ .KubeType }}
BASE_DIR={{ .BaseDir }}
LOCAL_GPU_ENABLE={{ .GpuEnable }}
LOCAL_GPU_SHARE={{ .GpuShare }}
    `)))
