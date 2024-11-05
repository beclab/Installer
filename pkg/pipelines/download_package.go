package pipelines

import (
	"path"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/download"
)

func DownloadInstallationPackage(opts *options.CliDownloadOptions) error {
	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetTerminusVersion(opts.Version)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	if manifest == "" {
		manifest = path.Join(runtime.GetInstallerDir(), "installation.manifest")
	}

	p := download.NewDownloadPackage(manifest, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("download package failed %v", err)
		return err
	}

	return nil
}
