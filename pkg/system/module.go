package system

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/action"
	corecommon "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

// ~ InstallDepsModule
type InstallDepsModule struct {
	common.KubeModule
}

func (m *InstallDepsModule) Init() {
	m.Name = "InstallDepsModule"

	installDeps := &task.RemoteTask{
		Name:  "InstallDeps",
		Hosts: m.Runtime.GetAllHosts(),
		Action: &action.Script{
			Name: "installDeps",
			File: corecommon.GreetingShell,
			Args: []string{constants.OsPlatform},
		},
	}

	m.Tasks = []task.Interface{
		installDeps,
	}
}
