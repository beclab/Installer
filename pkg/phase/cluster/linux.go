package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

type linuxInstallPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (l *linuxInstallPhaseBuilder) base() phase {
	m := []module.Module{
		&plugins.CopyEmbed{},
		&terminus.CheckPreparedModule{BaseDir: l.runtime.GetBaseDir(), Force: true},
		&terminus.TerminusUninstallScriptModule{},
	}

	return m
}

func (l *linuxInstallPhaseBuilder) installCluster() phase {
	kubeType := l.runtime.Arg.Kubetype
	if kubeType == common.K3s {
		return NewK3sCreateClusterPhase(l.runtime, l.manifestMap)
	} else {
		return NewCreateClusterPhase(l.runtime, l.manifestMap)
	}
}

func (l *linuxInstallPhaseBuilder) installGpuPlugin() phase {
	return []module.Module{
		&gpu.RestartK3sServiceModule{Skip: !(l.runtime.Arg.Kubetype == common.K3s && l.runtime.GetSystemInfo().IsWsl())},
		&gpu.InstallPluginModule{Skip: !(l.runtime.Arg.GPU.Enable || l.runtime.GetSystemInfo().IsWsl())},
	}
}

func (l *linuxInstallPhaseBuilder) installTerminus() phase {
	return []module.Module{
		&terminus.GetNATGatewayIPModule{},
		&terminus.InstallAccountModule{},
		&terminus.InstallSettingsModule{},
		&terminus.InstallOsSystemModule{},
		&terminus.InstallLauncherModule{},
		&terminus.InstallAppsModule{},
	}
}

func (l *linuxInstallPhaseBuilder) build() []module.Module {
	return l.base().
		addModule(l.installCluster()...).
		addModule(l.installGpuPlugin()...).
		addModule(l.installTerminus()...).
		addModule(&terminus.WelcomeModule{})
}
