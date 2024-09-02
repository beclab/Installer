package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
)

func DownloadInstallationWizard(opts *options.CliDownloadWizardOptions) error {
	arg := common.NewArgument()
	arg.SetTerminusVersion(opts.Version)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	home := runtime.GetHomeDir()
	baseDir := opts.BaseDir
	if baseDir == "" {
		baseDir = home + "/.terminus"
	}

	p := download.NewDownloadWizard(baseDir, opts.Md5sum, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("download wizard failed %v", err)
		return err
	}

	return nil
}
