package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/gpu"
)

type phase []module.Module

func (p phase) addModule(m ...module.Module) phase {
	return append(p, m...)
}

type gpuModuleBuilder func() []module.Module

func (m gpuModuleBuilder) withGPU(runtime *common.KubeRuntime) []module.Module {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsWsl() {
		if (&gpu.CheckWslGPU{}).CheckNvidiaSmiFileExists() {
			return m()
		}
	} else {
		return m()
	}
	return nil
}

type backupModuleBuilder func() []module.Module

func (m backupModuleBuilder) withBackup(runtime *common.KubeRuntime) []module.Module {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsLinux() {
		return m()
	}
	return nil
}
