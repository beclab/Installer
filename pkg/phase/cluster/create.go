package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

// only install kubesphere
func InitKube(args common.Argument, runtime *common.KubeRuntime) *pipeline.Pipeline {
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&plugins.GenerateCachedModule{},
		&plugins.CopyEmbed{},
	}

	var kubeModules []module.Module
	if args.Minikube {
		kubeModules = NewDarwinClusterPhase(runtime, manifestMap)
	} else {
		if runtime.Cluster.Kubernetes.Type == common.K3s {
			// FIXME:
			kubeModules = NewK3sCreateClusterPhase(runtime, nil)
		} else {
			kubeModules = NewCreateClusterPhase(runtime, nil)
		}

		kubeModules = append(kubeModules,
			&gpu.InstallDepsModule{Skip: !runtime.Arg.GPU.Enable},
			&gpu.RestartK3sServiceModule{Skip: !runtime.Arg.GPU.Enable},
			&gpu.RestartContainerdModule{Skip: !runtime.Arg.GPU.Enable},
			&gpu.InstallPluginModule{Skip: !runtime.Arg.GPU.Enable},
		)
	}
	m = append(m, kubeModules...)

	return &pipeline.Pipeline{
		Name:    "Initialize KubeSphere",
		Modules: m,
		Runtime: runtime,
	}
}

func CreateTerminus(args common.Argument, runtime *common.KubeRuntime) *pipeline.Pipeline {
	// TODO: the installation process needs to distinguish between macOS and Linux.
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	(&gpu.CheckWslGPU{}).Execute(runtime)

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		// FIXME: completely install supported
		&terminus.CheckPreparedModule{BaseDir: runtime.GetBaseDir(), Force: true},
		&terminus.TerminusPhaseStateModule{},
	}

	var kubeModules []module.Module
	if constants.OsType == common.Darwin {
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

	return &pipeline.Pipeline{
		Name:    "Install Terminus",
		Modules: m,
		Runtime: runtime,
	}
}
