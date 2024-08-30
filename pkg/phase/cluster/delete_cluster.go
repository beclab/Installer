package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/filesystem"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/loadbalancer"
	"bytetrade.io/web3os/installer/pkg/storage"
)

func DeleteMinikubePhase(args common.Argument, runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&kubesphere.DeleteCacheModule{},
		&kubesphere.DeleteMinikubeModule{},
		&filesystem.DeleteInstalledModule{},
	}
}

func DeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	var p = util.IsExist("/var/run/lock/.prepared")
	return []module.Module{
		&kubernetes.ResetClusterModule{},
		&container.UninstallContainerModule{Skip: p},
		&k3s.DeleteClusterModule{},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
		&kubesphere.DeleteCacheModule{},
		&storage.RemoveStorageModule{},
		&container.DeleteZfsMountModule{Skip: p},
		&k3s.UninstallK3sModule{},
		&filesystem.DeleteInstalledModule{},
	}
}
