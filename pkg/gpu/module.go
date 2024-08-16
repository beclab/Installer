package gpu

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallDepsModule struct {
	common.KubeModule
	Skip bool
}

func (m *InstallDepsModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDepsModule) Init() {
	m.Name = "InstallGPU"

	installCudaDeps := &task.LocalTask{
		Name: "InstallCudaKeyRing",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
			new(GPUEnablePrepare),
		},
		Action: new(InstallCudaDeps),
		// Parallel: false,
		Retry: 1,
	}

	installCudaDriver := &task.LocalTask{
		Name: "InstallNvidiaDriver",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
			new(GPUEnablePrepare),
		},
		Action: new(InstallCudaDriver),
		// Parallel: false,
		Retry: 1,
	}

	updateCudaSource := &task.LocalTask{
		Name: "UpdateNvidiaToolkitSource",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
			new(GPUEnablePrepare),
		},
		Action: new(UpdateCudaSource),
		// Parallel: false,
		Retry: 1,
	}

	installNvidiaContainerToolkit := &task.LocalTask{
		Name: "InstallNvidiaToolkit",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
			new(GPUEnablePrepare),
		},
		Action: new(InstallNvidiaContainerToolkit),
		// Parallel: false,
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		installCudaDeps,
		installCudaDriver,
		updateCudaSource,
		installNvidiaContainerToolkit,
	}
}

type RestartK3sServiceModule struct {
	common.KubeModule
}

func (m *RestartK3sServiceModule) Init() {
}
