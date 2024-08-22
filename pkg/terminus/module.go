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

type InstallWizardDownloadModule struct {
	common.KubeModule
	Version string
}

func (m *InstallWizardDownloadModule) Init() {
	m.Name = "DownloadInstallWizard"
	download := &task.LocalTask{
		Name: "DownloadInstallWizard",
		Action: &Download{
			version: m.Version,
		},
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		download,
	}
}

type CopyToWizardModule struct {
	common.KubeModule
}

func (m *CopyToWizardModule) Init() {
	m.Name = "CopyToInstallWizard"

	copyToWizard := &task.RemoteTask{
		Name:     "CopyToInstallWizard",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(CopyToWizard),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		copyToWizard,
	}
}

type DownloadFullInstallerModule struct {
	common.KubeModule
	Skip bool
}

func (r *DownloadFullInstallerModule) IsSkip() bool {
	return r.Skip
}

func (m *DownloadFullInstallerModule) Init() {
	m.Name = "DownloadFullInstaller"
}

type TidyPackageModule struct {
	common.KubeModule
	Skip bool
}

func (r *TidyPackageModule) IsSkip() bool {
	return r.Skip
}

func (m *TidyPackageModule) Init() {
	m.Name = "TidyInstallerPackage"

	tidyInstallerPackage := &task.LocalTask{
		Name:   "TidyInstallerPacker",
		Action: new(TidyInstallerPackage),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		tidyInstallerPackage,
	}
}

type PreparedModule struct {
	common.KubeModule
}

func (m *PreparedModule) Init() {
	m.Name = "PrepareFinished"

	prepareFinished := &task.LocalTask{
		Name:   "PrepareFinished",
		Action: new(PrepareFinished),
	}

	m.Tasks = []task.Interface{
		prepareFinished,
	}
}
