package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func NewDownloadWizard(baseDir, md5sum string, runtime *common.KubeRuntime) *pipeline.Pipeline {

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&terminus.InstallWizardDownloadModule{BaseDir: baseDir, Md5sum: md5sum, Version: runtime.Arg.TerminusVersion},
	}

	return &pipeline.Pipeline{
		Name:    "Download Terminus Installation Wizard",
		Modules: m,
		Runtime: runtime,
	}
}
