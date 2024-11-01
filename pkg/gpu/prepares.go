package gpu

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type GPUEnablePrepare struct {
	common.KubePrepare
}

func (p *GPUEnablePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsWsl() {
		return false, nil
	}

	if systemInfo.IsUbuntu() && systemInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
		return false, nil
	}
	return p.KubeConf.Arg.GPU.Enable, nil
}

type GPUSharePrepare struct {
	common.KubePrepare
}

func (p *GPUSharePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return p.KubeConf.Arg.GPU.Share || runtime.GetSystemInfo().IsWsl(), nil
}
