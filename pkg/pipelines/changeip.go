package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func ChangeIPPipeline(opt *options.ChangeIPOptions) error {
	terminusVersion := opt.Version
	kubeType := phase.GetKubeType()
	if terminusVersion == "" {
		terminusVersion, _ = phase.GetTerminusVersion()
	}

	var arg = common.NewArgument()
	arg.SetTerminusVersion(terminusVersion)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetConsoleLog("changeip.log", true)
	arg.SetKubeVersion(kubeType)
	arg.SetMinikubeProfile(opt.MinikubeProfile)
	arg.SetWSLDistribution(opt.WSLDistribution)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var p = cluster.ChangeIP(runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("failed to run change ip pipeline: %v", err)
		return err
	}

	return nil

}
