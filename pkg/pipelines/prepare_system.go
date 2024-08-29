package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(opts *options.CliPrepareSystemOptions) error {
	ksVersion, _, exists := kubernetes.CheckKubeExists()
	if exists {
		return fmt.Errorf("Kubernetes %s is already installed", ksVersion)
	}

	var terminusVersion = opts.Version // utils.GetTerminusVersion(version)

	var arg = common.NewArgument()
	arg.SetTerminusVersion(terminusVersion)
	arg.SetKubernetesVersion(common.K8s, common.DefaultKubernetesVersion)
	arg.SetProxy(opts.RegistryMirrors, opts.RegistryMirrors)
	arg.SetGPU(true, true)
	arg.SetStorage(createStorage(opts))
	arg.SetWSL(opts.WSL)

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

	runtime.Arg.SetBaseDir(baseDir)
	runtime.Arg.SetManifest(manifest)

	var p = system.PrepareSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}
