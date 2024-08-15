package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
)

func DebugCommand() error {
	var arg = common.NewArgument()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	m := []module.Module{
		&precheck.GreetingsModule{},
	}

	p := pipeline.Pipeline{
		Name:    "Debug Command",
		Modules: m,
		Runtime: runtime,
	}

	return p.Start()
}
