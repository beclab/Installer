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

	c.Tasks = []task.Interface{
		configWslConf,
		configWSLForwardRules,
		configWSLHostsAndDns,
	}
}

type InstallTerminusModule struct {
	common.KubeModule
}

func (i *InstallTerminusModule) Init() {
	i.Name = "InstallTerminus"
	i.Tasks = []task.Interface{
		&task.LocalTask{
			Name:   "InstallTerminus",
			Action: &InstallTerminus{},
		},
	}
}
