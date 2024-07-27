package patch

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallDepsModule struct {
	common.KubeModule
}

func (m *InstallDepsModule) Init() {
	m.Name = "InstallDeps"

	patchOs := &task.RemoteTask{
		Name:   "PatchOs",
		Hosts:  m.Runtime.GetAllHosts(),
		Action: new(PatchTask),
		Retry:  0,
	}

	installSocat := &task.RemoteTask{
		Name:     "InstallSocat",
		Hosts:    m.Runtime.GetAllHosts(),
		Prepare:  &CheckDepsPrepare{Command: common.CommandSocat},
		Action:   new(SocatTask),
		Parallel: false,
		Retry:    0,
	}

	installConntrack := &task.RemoteTask{
		Name:     "InstallConntrack",
		Hosts:    m.Runtime.GetAllHosts(),
		Prepare:  &CheckDepsPrepare{Command: common.CommandConntrack},
		Action:   new(ConntrackTask),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		patchOs,
		installSocat,
		installConntrack,
	}
}
