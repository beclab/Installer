package windows

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type DownloadImageModule struct {
	common.KubeModule
}

func (d *DownloadImageModule) Init() {
	d.Name = "DownloadWslImage"
	downloadWslImage := &task.LocalTask{
		Name:   "DownloadWslImage",
		Action: new(DownloadImage),
	}
	d.Tasks = []task.Interface{
		downloadWslImage,
	}
}

type ImportImageModule struct {
	common.KubeModule
}

func (i *ImportImageModule) Init() {
	i.Name = "ImportTerminusDistro"
	importWslImage := &task.LocalTask{
		Name:   "ImportTerminusDistro",
		Action: new(ImportImage),
	}

	i.Tasks = []task.Interface{
		importWslImage,
	}
}

type ConfigWslModule struct {
	common.KubeModule
}

func (c *ConfigWslModule) Init() {
	c.Name = "ConfigWslConfig"

	configWSLForwardRules := &task.LocalTask{
		Name:   "ConfigWslConfig",
		Action: new(ConfigWSLForwardRules),
	}

	configWSLHostsAndDns := &task.LocalTask{
		Name:   "ConfigWslHostsAndDns",
		Action: new(ConfigWSLHostsAndDns),
	}

	c.Tasks = []task.Interface{
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
			Action: new(InstallTerminus),
		},
	}
}
