package cluster

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func CreateTerminus(runtime *common.KubeRuntime) *pipeline.Pipeline {
	// TODO: the installation process needs to distinguish between macOS and Linux.
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	(&gpu.CheckWslGPU{}).Execute(runtime)

	m := []module.Module{
		&plugins.CopyEmbed{},
		&terminus.CheckPreparedModule{BaseDir: runtime.GetBaseDir(), Force: true},
		&terminus.TerminusUninstallScriptModule{},
	}

	var kubeModules []module.Module
	if runtime.GetSystemInfo().IsDarwin() {
		kubeModules = NewDarwinClusterPhase(runtime, manifestMap)
	} else {
		if runtime.Cluster.Kubernetes.Type == common.K3s {
			kubeModules = NewK3sCreateClusterPhase(runtime, manifestMap)
		} else {
			kubeModules = NewCreateClusterPhase(runtime, manifestMap)
		}
		kubeModules = append(kubeModules,
			&gpu.InstallPluginModule{Skip: !runtime.Arg.GPU.Enable},
		)
	}
	m = append(m, kubeModules...)
	m = append(m, terminus.GenerateTerminusComponentsModules(runtime, manifestMap)...)
	m = append(m, &terminus.InstalledModule{})

	return &pipeline.Pipeline{
		Name:    "Install Terminus",
		Modules: m,
		Runtime: runtime,
	}
}
