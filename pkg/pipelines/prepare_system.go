package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(opts *options.CliPrepareSystemOptions) error {
	ksVersion, _, exists := kubernetes.CheckKubeExists()
	if exists {
		return fmt.Errorf("Kubernetes %s is already installed", ksVersion)
	}

	logger.Infof("prepare args, baseDir: %s, version: %s, mirrors: %s",
		opts.BaseDir, opts.Version, opts.RegistryMirrors,
	)

	var arg = common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetTerminusVersion(opts.Version)
	arg.SetKubernetesVersion(opts.KubeType, "")
	arg.SetRegistryMirrors(opts.RegistryMirrors)
	arg.SetStorage(createStorage(opts))
	arg.SetReverseProxy()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	manifest := opts.Manifest
	home := runtime.GetHomeDir()
	if manifest == "" {
		manifest = home + "/.terminus/installation.manifest"
	}

	runtime.Arg.SetManifest(manifest)

	var p = system.PrepareSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}
