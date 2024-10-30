package patch

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type InstallDepsModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *InstallDepsModule) Init() {
	logger.InfoInstallationProgress("installing and configuring OS dependencies ...")
	m.Name = "InstallDeps"

	patchOs := &task.RemoteTask{
		Name:   "PatchOs",
		Hosts:  m.Runtime.GetAllHosts(),
		Action: new(PatchTask),
		Retry:  0,
	}

	enableSSHTask := &task.RemoteTask{
		Name:     "EnableSSH",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(EnableSSHTask),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		patchOs,
		enableSSHTask,
	}
}
