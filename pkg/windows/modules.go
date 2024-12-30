package windows

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallWSLModule struct {
	common.KubeModule
}

func (u *InstallWSLModule) Init() {
	u.Name = "InstallWSL"

	downloadAppxPackage := &task.LocalTask{
		Name:   "InitAppxPackage",
		Action: &AddAppxPackage{},
	}

	updateWSL := &task.LocalTask{
		Name:   "UpdateWSL",
		Action: &UpdateWSL{},
	}

	u.Tasks = []task.Interface{
		downloadAppxPackage,
		updateWSL,
	}
}

type InstallWSLUbuntuDistroModule struct {
	common.KubeModule
}

func (i *InstallWSLUbuntuDistroModule) Init() {
	i.Name = "InstallWSLUbuntuDistro"

	installWSLDistro := &task.LocalTask{
		Name:   "InstallWSLDistro",
		Action: &InstallWSLDistro{},
		Retry:  1,
	}

	i.Tasks = []task.Interface{
		installWSLDistro,
	}
}

type ConfigWslModule struct {
	common.KubeModule
}

func (c *ConfigWslModule) Init() {
	c.Name = "ConfigWslConfig"

	configWslConf := &task.LocalTask{
		Name:   "ConfigWslConf",
		Action: &ConfigWslConf{},
	}

	configWSLForwardRules := &task.LocalTask{
		Name:   "ConfigWslConfig",
		Action: &ConfigWSLForwardRules{},
	}

	configWSLHostsAndDns := &task.LocalTask{
		Name:   "ConfigWslHostsAndDns",
		Action: &ConfigWSLHostsAndDns{},
	}

	configWindowsFirewallRule := &task.LocalTask{
		Name:   "ConfigFirewallRule",
		Action: &ConfigWindowsFirewallRule{},
	}

	c.Tasks = []task.Interface{
		configWslConf,
		configWSLForwardRules,
		configWSLHostsAndDns,
		configWindowsFirewallRule,
	}
}

type InstallTerminusModule struct {
	common.KubeModule
}

func (i *InstallTerminusModule) Init() {
	i.Name = "InstallOlares"
	i.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "InstallOlares",
			Action: &InstallTerminus{},
		},
	}
}

type UninstallOlaresModule struct {
	common.KubeModule
}

func (u *UninstallOlaresModule) Init() {
	u.Name = "UninstallOlares"
	u.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "UninstallOlares",
			Action: &UninstallOlares{},
		},
		&task.LocalTask{
			Name:   "RemoveFirewallRule",
			Action: &RemoveFirewallRule{},
		},
		&task.LocalTask{
			Name:   "RemovePortProxy",
			Action: &RemovePortProxy{},
		},
	}
}
