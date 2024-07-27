package kk

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallKkModule struct {
	common.KubeModule
}

func (m *InstallKkModule) Init() {
	m.Name = "InstallKkModule"
	m.Desc = "Install kubekey"

	chmodKk := &task.RemoteTask{
		Name:   "ChmodKk",
		Desc:   "Chmod kubekey",
		Hosts:  m.Runtime.GetAllHosts(),
		Action: new(ChmodKk),
	}

	installKk := &task.RemoteTask{
		Name:   "InstallKk",
		Desc:   "Install kubekey",
		Hosts:  m.Runtime.GetAllHosts(),
		Action: new(ExecuteKk),
	}

	m.Tasks = []task.Interface{chmodKk, installKk}
}
