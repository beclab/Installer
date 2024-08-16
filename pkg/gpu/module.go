package gpu

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

type InstallDepsModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallDepsModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDepsModule) Init() {
	m.Name = "InstallGPU"

	copyEmbedGpuFiles := &task.LocalTask{
		Name: "CopyFiles",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
		},
		Action: new(CopyEmbedGpuFiles),
		// Parallel: false,
		Retry: 1,
	}

	installCudaDeps := &task.LocalTask{
		Name: "InstallCudaKeyRing",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
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
		},
		Action: new(InstallNvidiaContainerToolkit),
		// Parallel: false,
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		copyEmbedGpuFiles,
		installCudaDeps,
		installCudaDriver,
		updateCudaSource,
		installNvidiaContainerToolkit,
	}
}

type RestartK3sServiceModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *RestartK3sServiceModule) IsSkip() bool {
	return m.Skip
}

func (m *RestartK3sServiceModule) Init() {
	m.Name = "RestartK3sService"

	patchK3sDriver := &task.LocalTask{
		Name: "PatchK3sDriver",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
			new(PatchWslK3s),
		},
		Action: new(PatchK3sDriver),
		// Parallel: false,
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		patchK3sDriver,
	}
}

type RestartContainerdModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *RestartContainerdModule) IsSkip() bool {
	return m.Skip
}

func (m *RestartContainerdModule) Init() {
	m.Name = "RestartContainerd"

	restartContainerd := &task.LocalTask{
		Name: "RestartContainerd",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
		},
		Action: new(RestartContainerd),
		// Parallel: false,
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		restartContainerd,
	}
}

type InstallPluginModule struct {
	common.KubeModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallPluginModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallPluginModule) Init() {
	m.Name = "InstallPlugin"

	installPlugin := &task.LocalTask{
		Name: "InstallPlugin",
		// Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			// new(common.OnlyFirstMaster),
		},
		Action: new(InstallPlugin),
		// Parallel: false,
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		installPlugin,
	}
}
