package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
)

func DownloadInstallationPackage(opts *options.CliDownloadOptions) error {
	arg := common.NewArgument()
	arg.SetTerminusVersion(opts.Version)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	home := runtime.GetHomeDir()
	if manifest == "" {
		manifest = home + "/.terminus/installation.manifest"
	}

	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = home + "/.terminus"
	}

	p := download.NewDownloadPackage(manifest, baseDir, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("download package failed %v", err)
		return err
	}

	return nil
}
