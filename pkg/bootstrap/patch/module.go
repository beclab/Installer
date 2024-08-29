package patch

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type InstallDepsModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *InstallDepsModule) Init() {
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

	installSocat := &task.RemoteTask{
		Name:    "InstallSocat",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckDepsPrepare{Command: common.CommandSocat},
		Action: &SocatTask{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	installConntrack := &task.RemoteTask{
		Name:    "InstallConntrack",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckDepsPrepare{Command: common.CommandConntrack},
		Action: &ConntrackTask{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		patchOs,
		enableSSHTask,
		installSocat,
		installConntrack,
	}
}
