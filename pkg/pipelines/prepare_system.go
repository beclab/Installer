package pipelines

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/phase/system"
)

func PrepareSystemPipeline(version string, proxy string) error {
	ksVersion, _, exists := kubernetes.CheckKubeExists()
	if exists {
		return fmt.Errorf("Kubernetes %s is already installed", ksVersion)
	}

	var terminusVersion = version // utils.GetTerminusVersion(version)

	var arg = common.NewArgument()
	arg.SetTerminusVersion(terminusVersion)
	arg.SetKubernetesVersion(common.K8s, common.DefaultKubernetesVersion)
	arg.SetProxy(proxy, proxy)
	arg.SetGPU(true, true)

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
