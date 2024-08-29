package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/download"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func NewDownloadPackage(mainifest, baseDir string, runtime *common.KubeRuntime) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&terminus.InstallWizardDownloadModule{BaseDir: baseDir, Version: runtime.Arg.TerminusVersion},
		&download.PackageDownloadModule{Manifest: mainifest, BaseDir: baseDir},
	}

	return &pipeline.Pipeline{
		Name:    "Download Terminus Installation Package",
		Modules: m,
		Runtime: runtime,
	}
}
