package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(version string) error {
	ksVersion, _, exists := kubernetes.CheckKubeExists()
	if exists {
		return fmt.Errorf("Kubernetes %s is already installed", ksVersion)
	}

	var arg = common.NewArgument()
	arg.SetTerminusVersion(version)
	arg.SetKubernetesVersion(common.K8s, common.DefaultKubernetesVersion)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var p = system.PrepareSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return err
	}

	return nil
}
