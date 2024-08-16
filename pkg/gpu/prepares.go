package gpu

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type GPUEnablePrepare struct {
	common.KubePrepare
}

func (p *GPUEnablePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	if p.KubeConf.Arg.WSL {
		return false, nil
	}
	if constants.OsPlatform == common.Ubuntu && strings.Contains(constants.OsVersion, "24.") {
		return false, nil
	}
	return p.KubeConf.Arg.GPU.Enable, nil
}

type PatchWslK3s struct {
	common.KubePrepare
}

func (p *PatchWslK3s) PreCheck(runtime connector.Runtime) (bool, error) {
	if p.KubeConf.Cluster.Kubernetes.Type == common.K3s {
		if p.KubeConf.Arg.WSL {
			return true, nil
		}
	}
	return false, nil
}

type GPUSharePrepare struct {
	common.KubePrepare
}

func (p *GPUSharePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	return p.KubeConf.Arg.GPU.Share, nil
}
