package download

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/download"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func NewCheckDownload(mainifest, baserdir string, runtime *common.KubeRuntime) *pipeline.Pipeline {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&download.PackageDownloadModule{Manifest: mainifest, BaseDir: baserdir},
	}

	return &pipeline.Pipeline{
		Name:    "Check Downloaded Terminus Installation Package",
		Modules: m,
		Runtime: runtime,
	}
}
