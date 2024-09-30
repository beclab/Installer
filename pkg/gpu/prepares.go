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

	if systemInfo.IsUbuntu() && systemInfo.IsUbuntuVersionAbove(connector.UbuntuAbove24) {
		return false, nil
	}
	return p.KubeConf.Arg.GPU.Enable, nil
}

type GPUSharePrepare struct {
	common.KubePrepare
}

func (p *GPUSharePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return p.KubeConf.Arg.GPU.Share, nil
}
