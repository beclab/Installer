package gpu

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type InstallDepsModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallDepsModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDepsModule) Init() {
	m.Name = "InstallGPU"

	// copyEmbedGpuFiles := &task.RemoteTask{
	// 	Name:  "CopyFiles",
	// 	Hosts: m.Runtime.GetHostsByRole(common.Master),
	// 	Prepare: &prepare.PrepareCollection{
	// 		new(common.OnlyFirstMaster),
	// 	},
	// 	Action:   new(CopyEmbedGpuFiles),
	// 	Parallel: false,
	// 	Retry:    1,
	// }

	installCudaDeps := &task.RemoteTask{
		Name:  "InstallCudaKeyRing",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&common.OsType{
				OsType: common.Ubuntu, // ! only ubuntu
			},
			&common.OsVersion{
				OsVersion: map[string]bool{
					"20.04": true,
					"22.04": true,
					"24.04": false,
				},
			},
		},
		Action: &InstallCudaDeps{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	installCudaDriver := &task.RemoteTask{ // not for WSL
		Name:  "InstallNvidiaDriver",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&common.Skip{
				Not: !m.KubeConf.Arg.WSL,
			},
		},
		Action:   new(InstallCudaDriver),
		Parallel: false,
		Retry:    1,
	}

	updateCudaSource := &task.RemoteTask{
		Name:  "UpdateNvidiaToolkitSource",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action: &UpdateCudaSource{
			ManifestAction: manifest.ManifestAction{
				Manifest: m.Manifest,
				BaseDir:  m.BaseDir,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	installNvidiaContainerToolkit := &task.RemoteTask{
		Name:  "InstallNvidiaToolkit",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(InstallNvidiaContainerToolkit),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		// copyEmbedGpuFiles,
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

	patchK3sDriver := &task.RemoteTask{
		Name:  "PatchK3sDriver",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(PatchWslK3s),
		},
		Action:   new(PatchK3sDriver),
		Parallel: false,
		Retry:    1,
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

	restartContainerd := &task.RemoteTask{
		Name:  "RestartContainerd",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(RestartContainerd),
		Parallel: false,
		Retry:    1,
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

	installPlugin := &task.RemoteTask{
		Name:  "InstallPlugin",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(InstallPlugin),
		Parallel: false,
		Retry:    1,
	}

	checkGpuState := &task.RemoteTask{
		Name:  "CheckGPUState",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(CheckGpuStatus),
		Parallel: false,
		Retry:    50,
		Delay:    10 * time.Second,
	}

	installGPUShared := &task.RemoteTask{
		Name:  "InstallGPUShared",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(GPUSharePrepare),
		},
		Action:   new(InstallGPUShared),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installPlugin,
		checkGpuState,
		installGPUShared,
	}
}
