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
		&terminus.CheckPreparedModule{Force: true},
		&terminus.OlaresUninstallScriptModule{},
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
		&gpu.RestartK3sServiceModule{Skip: !(l.runtime.Arg.Kubetype == common.K3s)},
		&gpu.InstallPluginModule{Skip: !l.runtime.Arg.GPU.Enable},
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
		addModule(gpuModuleBuilder(func() []module.Module {
			return l.installGpuPlugin()
		}).withGPU(l.runtime)...).
		addModule(l.installTerminus()...).
		addModule(&terminus.WelcomeModule{}).
		addModule(&terminus.InstalledModule{})
}
