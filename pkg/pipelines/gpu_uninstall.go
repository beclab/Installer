package pipelines

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
)

func UninstallGpuDrivers() error {

	arg := common.NewArgument()
	arg.SetConsoleLog("gpuuninstall.log", true)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	p := &pipeline.Pipeline{
		Name:    "UninstallGpuDrivers",
		Runtime: runtime,
		Modules: []module.Module{
			&gpu.NodeUnlabelingModule{},
			&gpu.UninstallCudaModule{},
			&gpu.RestartContainerdModule{},
		},
	}

	return p.Start()

}
