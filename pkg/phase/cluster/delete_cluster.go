package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
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
	return []module.Module{
		&kubernetes.ResetClusterModule{},
		&k3s.DeleteClusterModule{},
		&container.UninstallContainerModule{Skip: !runtime.Arg.DeleteCRI},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
		&kubesphere.DeleteCacheModule{},
		&storage.RemoveStorageModule{},
		&k3s.UninstallK3sModule{},
		&filesystem.DeleteInstalledModule{},
	}
}
