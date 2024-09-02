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

func DeleteClusterPhase(baseDir, phase string, runtime *common.KubeRuntime) []module.Module {
	var deleteVipModule = !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()
	var res []module.Module

	switch phase {
	case "download":
		res = append(res,
			&filesystem.DeleteInstalledModule{
				BaseDir: baseDir,
			},
		)
		fallthrough
	case "prepare":
		res = append(res,
			&container.UninstallContainerModule{},
			&storage.RemoveStorageModule{},
			&container.DeleteZfsMountModule{},
		)
		fallthrough
	case "install":
		res = append(res,
			&k3s.UninstallK3sModule{},
			&loadbalancer.DeleteVIPModule{Skip: deleteVipModule},
			&certs.UninstallAutoRenewCertsModule{},
			&os.ClearOSEnvironmentModule{},
			&k3s.DeleteClusterModule{},
			&kubernetes.ResetClusterModule{},
		)
	}

	// res = []module.Module{
	// 	&kubernetes.ResetClusterModule{},
	// 	&container.UninstallContainerModule{Skip: p},
	// 	&k3s.DeleteClusterModule{},
	// 	&os.ClearOSEnvironmentModule{},
	// 	&certs.UninstallAutoRenewCertsModule{},
	// 	&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
	// 	&kubesphere.DeleteCacheModule{},
	// 	&storage.RemoveStorageModule{},
	// 	&k3s.UninstallK3sModule{},
	// 	&filesystem.DeleteInstalledModule{},
	// }

	reverseModules(res)

	return res
}

func reverseModules(s []module.Module) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
