package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/loadbalancer"
	"bytetrade.io/web3os/installer/pkg/storage"
)

func DeleteMinikubePhase(args common.Argument, runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&precheck.GreetingsModule{},
		&kubesphere.DeleteCacheModule{},
		&kubesphere.DeleteMinikubeModule{},
	}
}

func DeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	var kubeModule []module.Module
	switch runtime.Cluster.Kubernetes.Type {
	case common.K3s:
		kubeModule = newK3sDeleteClusterPhase(runtime)
	case common.Kubernetes:
		kubeModule = newK8sDeleteClusterPhase(runtime)
	}
	kubeModule = append(kubeModule,
		&kubesphere.DeleteCacheModule{},
		&storage.RemoveStorageModule{},
		&storage.RemoveMountModule{},
		&k3s.UninstallK3sModule{},
	)

	return kubeModule
}

func newK8sDeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&precheck.GreetingsModule{},
		&kubernetes.ResetClusterModule{},
		&container.UninstallContainerModule{Skip: !runtime.Arg.DeleteCRI},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
	}
}

func newK3sDeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&precheck.GreetingsModule{},
		&k3s.DeleteClusterModule{},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
	}
}
