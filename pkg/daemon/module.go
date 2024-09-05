package daemon

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type InstallTerminusdBinaryModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (i *InstallTerminusdBinaryModule) Init() {
	i.Name = "InstallTerminusdBinaryModule"
	i.Desc = "Install terminusd"

	install := &task.RemoteTask{
		Name:  "InstallTerminusdBinary",
		Desc:  "Install terminusd using binary",
		Hosts: i.Runtime.GetHostsByRole(common.K8s),
		Action: &InstallTerminusdBinary{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  i.BaseDir,
				Manifest: i.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	generateEnv := &task.RemoteTask{
		Name:     "GenerateTerminusdEnv",
		Desc:     "Generate terminus service env",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GenerateTerminusdServiceEnv),
		Parallel: false,
		Retry:    1,
	}

	generateService := &task.RemoteTask{
		Name:     "GenerateTerminusdService",
		Desc:     "Generate terminus service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GenerateTerminusdService),
		Parallel: false,
		Retry:    1,
	}

	enableService := &task.RemoteTask{
		Name:     "EnableTerminusdService",
		Desc:     "enable terminus service",
		Hosts:    i.Runtime.GetHostsByRole(common.K8s),
		Action:   new(EnableTerminusdService),
		Parallel: false,
		Retry:    1,
	}

	i.Tasks = []task.Interface{
		install,
		generateEnv,
		generateService,
		enableService,
	}
}
