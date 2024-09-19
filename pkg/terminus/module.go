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
			Version: m.Version,
		},
		Retry: 1,
	}

	m.Tasks = []task.Interface{
		download,
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

type PreparedModule struct {
	common.KubeModule
	BaseDir string
}

func (m *PreparedModule) Init() {
	m.Name = "PrepareFinished"

	prepareFinished := &task.LocalTask{
		Name: "PrepareFinished",
		Action: &PrepareFinished{
			BaseDir: m.BaseDir,
		},
	}

	m.Tasks = []task.Interface{
		prepareFinished,
	}
}

type CheckPreparedModule struct {
	common.KubeModule
	BaseDir string
	Force   bool
}

func (m *CheckPreparedModule) Init() {
	m.Name = "CheckPrepared"

	checkPrepared := &task.LocalTask{
		Name:   "CheckPrepared",
		Action: &CheckPepared{Force: m.Force, BaseDir: m.BaseDir},
	}

	m.Tasks = []task.Interface{
		checkPrepared,
	}
}

type TerminusUninstallScriptModule struct {
	common.KubeModule
}

func (m *TerminusUninstallScriptModule) Init() {
	m.Name = "GenerateTerminusUninstallScript"

	uninstallScript := &task.LocalTask{
		Name:   "GenerateTerminusUninstallScript",
		Action: &GenerateTerminusUninstallScript{},
	}

	m.Tasks = []task.Interface{
		uninstallScript,
	}
}

type TerminusPhaseStateModule struct {
	common.KubeModule
}

func (m *TerminusPhaseStateModule) Init() {
	m.Name = "GeneratePhaseState"

	installedState := &task.LocalTask{
		Name:   "GenerateInstalledPhase",
		Action: &GenerateInstalledPhaseState{},
	}

	m.Tasks = []task.Interface{
		installedState,
	}
}
