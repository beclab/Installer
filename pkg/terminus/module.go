package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type SetupWs struct {
	common.KubeModule
}

func (m *SetupWs) Init() {
	m.Name = "SetupWs"

	setUserInfo := &task.RemoteTask{
		Name:  "CreateStorageDir",
		Hosts: m.Runtime.GetAllHosts(),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(NotEqualDesiredVersion),
			new(CheckAwsHost),
		},
		Action:   new(SetUserInfo),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		setUserInfo,
	}
}
