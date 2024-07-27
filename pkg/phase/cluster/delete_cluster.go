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
	"bytetrade.io/web3os/installer/pkg/loadbalancer"
)

func NewK8sDeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&precheck.GreetingsModule{},
		&kubernetes.ResetClusterModule{},
		&container.UninstallContainerModule{Skip: !runtime.Arg.DeleteCRI},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
	}
}

func NewK3sDeleteClusterPhase(runtime *common.KubeRuntime) []module.Module {
	return []module.Module{
		&precheck.GreetingsModule{},
		&k3s.DeleteClusterModule{},
		&os.ClearOSEnvironmentModule{},
		&certs.UninstallAutoRenewCertsModule{},
		&loadbalancer.DeleteVIPModule{Skip: !runtime.Cluster.ControlPlaneEndpoint.IsInternalLBEnabledVip()},
	}
}
