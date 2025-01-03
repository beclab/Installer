package gpu

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type InstallDriversModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallDriversModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallDriversModule) Init() {
	m.Name = "InstallGPUDriver"

	installCudaDeps := &task.RemoteTask{
		Name:  "InstallCudaKeyRing",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&CudaNotInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(NvidiaGraphicsCard),
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
			&CudaNotInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(NvidiaGraphicsCard),
		},
		Action:   new(InstallCudaDriver),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installCudaDeps,
		installCudaDriver,
	}
}

type InstallContainerToolkitModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool // enableGPU && ubuntuVersionSupport
}

func (m *InstallContainerToolkitModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallContainerToolkitModule) Init() {
	m.Name = "InstallContainerToolkit"

	updateCudaSource := &task.RemoteTask{
		Name:  "UpdateNvidiaToolkitSource",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
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
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(ContainerdInstalled),
		},
		Action:   new(InstallNvidiaContainerToolkit),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
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
			new(ContainerdInstalled),
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

	// update node with gpu labels, to make plugins enabled
	updateNode := &task.RemoteTask{
		Name:  "UpdateNode",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
		},
		Action:   new(UpdateNodeLabels),
		Parallel: false,
		Retry:    1,
	}

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
		updateNode,
		installPlugin,
		checkGpuState,
		installGPUShared,
	}
}

type GetCudaVersionModule struct {
	common.KubeModule
}

func (g *GetCudaVersionModule) Init() {
	g.Name = "GetCudaVersion"

	getCudaVersion := &task.LocalTask{
		Name:   "GetCudaVersion",
		Action: new(GetCudaVersion),
	}

	g.Tasks = []task.Interface{
		getCudaVersion,
	}
}

type NodeLabelingModule struct {
	common.KubeModule
}

func (l *NodeLabelingModule) Init() {
	l.Name = "NodeLabeling"

	updateNode := &task.RemoteTask{
		Name:  "UpdateNode",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(K8sNodeInstalled),
		},
		Action:   new(UpdateNodeLabels),
		Parallel: false,
		Retry:    1,
	}

	restartPlugin := &task.RemoteTask{
		Name:  "RestartPlugin",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(K8sNodeInstalled),
		},
		Action:   new(RestartPlugin),
		Parallel: false,
		Retry:    1,
	}

	l.Tasks = []task.Interface{
		updateNode,
		restartPlugin,
	}
}

type NodeUnlabelingModule struct {
	common.KubeModule
}

func (l *NodeUnlabelingModule) Init() {
	l.Name = "NodeUnlabeling"

	removeNode := &task.RemoteTask{
		Name:  "RemoveNodeLabels",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(K8sNodeInstalled),
		},
		Action:   new(RemoveNodeLabels),
		Parallel: false,
		Retry:    1,
	}

	restartPlugin := &task.RemoteTask{
		Name:  "RestartPlugin",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
			new(K8sNodeInstalled),
		},
		Action:   new(RestartPlugin),
		Parallel: false,
		Retry:    1,
	}

	l.Tasks = []task.Interface{
		removeNode,
		restartPlugin,
	}
}

type UninstallCudaModule struct {
	common.KubeModule
}

func (l *UninstallCudaModule) Init() {
	l.Name = "UninstallCuda"

	uninstallCuda := &task.RemoteTask{
		Name:  "UninstallCuda",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			&CudaInstalled{
				CudaCheckTask: precheck.CudaCheckTask{
					SupportedCudaVersion: common.DefaultCudaVersion,
				},
			},
		},
		Action:   new(UninstallNvidiaDrivers),
		Parallel: false,
		Retry:    1,
	}

	removeRuntime := &task.RemoteTask{
		Name:  "RemoveRuntime",
		Hosts: l.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(ContainerdInstalled),
		},
		Action: new(RemoveContainerRuntimeConfig),
	}

	l.Tasks = []task.Interface{
		uninstallCuda,
		removeRuntime,
	}

}
