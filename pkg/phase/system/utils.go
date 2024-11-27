package system

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/module"
)

func isGpuSupportOs(runtime *common.KubeRuntime) bool {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsUbuntu() && (systemInfo.IsUbuntuVersionEqual(connector.Ubuntu20) || systemInfo.IsUbuntuVersionEqual(connector.Ubuntu22)) {
		return true
	}

	return false
}

type phase []module.Module

func (p phase) addModule(m ...module.Module) phase {
	return append(p, m...)
}

type cloudModuleBuilder func() []module.Module

func (m cloudModuleBuilder) withCloud(runtime *common.KubeRuntime) []module.Module {
	if runtime.Arg.IsCloudInstance {
		return m()
	}

	return nil
}

func (m cloudModuleBuilder) withoutCloud(runtime *common.KubeRuntime) []module.Module {
	if !runtime.Arg.IsCloudInstance {
		return m()
	}

	return nil
}

type storageModuleBuilder func() []module.Module

func (m storageModuleBuilder) withJuiceFS(runtime *common.KubeRuntime) []module.Module {
	// if juicefs is enabled
	// install redis/minio/juicefs
	if runtime.Arg.WithJuiceFS {
		return m()
	}
	// use local disk storage
	// so nothing need to be done
	return nil
}

type gpuModuleBuilder func() []module.Module

func (m gpuModuleBuilder) withGPU(runtime *common.KubeRuntime) []module.Module {
	if runtime.Arg.GPU.Enable && isGpuSupportOs(runtime) {
		return m()
	}

	return nil
}

type terminusBoxModuleBuilder func() []module.Module

func (m terminusBoxModuleBuilder) inBox(runtime *common.KubeRuntime) []module.Module {
	if runtime.GetSystemInfo().IsDarwin() {
		return nil
	}
	return m()
}
