package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

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

type PreparedModule struct {
	common.KubeModule
}

func (m *PreparedModule) Init() {
	m.Name = "PrepareFinished"

	prepareFinished := &task.LocalTask{
		Name:   "PrepareFinished",
		Action: &PrepareFinished{},
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

type InstallComponentsInClusterModule struct {
	common.KubeModule
}

type GetNATGatewayIPModule struct {
	common.KubeModule
}

func (m *GetNATGatewayIPModule) Init() {
	m.Name = "GetNATGatewayIP"

	getNATGatewayIP := &task.LocalTask{
		Name:   "GetNATGatewayIP",
		Action: new(GetNATGatewayIP),
	}

	m.Tasks = []task.Interface{
		getNATGatewayIP,
	}
}

func GenerateTerminusComponentsModules(runtime connector.Runtime, manifestMap manifest.InstallationManifest) []module.Module {
	var modules []module.Module
	baseModules := []module.Module{
		&GetNATGatewayIPModule{},
		&InstallAccountModule{},
		&InstallSettingsModule{},
		&InstallOsSystemModule{},
		&InstallLauncherModule{},
		&InstallAppsModule{},
	}
	modules = append(modules, baseModules...)

	si := runtime.GetSystemInfo()
	switch {
	case si.IsDarwin():
	default:
		modules = append(
			modules,
			&InstallVeleroModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: manifestMap,
					BaseDir:  runtime.GetBaseDir(),
				},
			})
	}

	modules = append(modules, &WelcomeModule{})

	return modules
}

type InstalledModule struct {
	common.KubeModule
}

func (m *InstalledModule) Init() {
	m.Name = "InstallFinished"

	installedState := &task.LocalTask{
		Name:   "InstallFinished",
		Action: &InstallFinished{},
	}

	m.Tasks = []task.Interface{
		installedState,
	}
}

type DeleteWizardFilesModule struct {
	common.KubeModule
}

func (d *DeleteWizardFilesModule) Init() {
	d.Name = "DeleteWizardFiles"

	deleteWizardFiles := &task.LocalTask{
		Name:   "DeleteWizardFiles",
		Action: &DeleteWizardFiles{},
	}

	d.Tasks = []task.Interface{
		deleteWizardFiles,
	}
}
