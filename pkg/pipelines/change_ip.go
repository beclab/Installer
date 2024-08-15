package pipelines

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func ChangeClusterIPPipeline(opt *options.ChangeIPOptions) error {
	var arg = common.NewArgument()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var m = cluster.ChangeIPPhase(runtime)

	p := pipeline.Pipeline{
		Name:    "Change IP",
		Runtime: runtime,
		Modules: m,
	}

	if err := p.Start(); err != nil {
		logger.Errorf("change ip failed: %v", err)
		return err
	}

	return nil
}
