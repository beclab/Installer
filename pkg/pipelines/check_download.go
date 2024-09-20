package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
)

func CheckDownloadInstallationPackage(opts *options.CliDownloadOptions) error {
	arg := common.NewArgument()
	arg.SetTerminusVersion(opts.Version)
	arg.SetBaseDir(opts.BaseDir)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	home := runtime.GetHomeDir()
	if manifest == "" {
		manifest = home + "/.terminus/installation.manifest"
	}

	baseDir := runtime.GetBaseDir()

	p := download.NewCheckDownload(manifest, baseDir, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("check download package failed %v", err)
		return err
	}

	return nil
}
