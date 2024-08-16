package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/storage"
)

// only install kubesphere
func InitKube(args common.Argument, runtime *common.KubeRuntime) *pipeline.Pipeline {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&plugins.GenerateCachedModule{},
		&plugins.CopyManifestModule{},
		&plugins.CopyEmbed{},
	}

	var kubeModules []module.Module
	if args.Minikube {
		kubeModules = NewDarwinClusterPhase(runtime)
	} else {
		if runtime.Cluster.Kubernetes.Type == common.K3s {
			kubeModules = NewK3sCreateClusterPhase(runtime)
		} else {
			kubeModules = NewCreateClusterPhase(runtime)
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
	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		&precheck.PreCheckOsModule{},                            // precheck_os()
		&patch.InstallDepsModule{},                              // install_deps
		&os.ConfigSystemModule{},                                // config_system
		&storage.InitStorageModule{Skip: !args.IsCloudInstance}, // todo skip when WSL
		&storage.InstallMinioModule{Skip: args.Storage.StorageType != common.Minio},
		&storage.InstallRedisModule{},
		&storage.InstallJuiceFsModule{},
		&plugins.GenerateCachedModule{},
		&plugins.CopyManifestModule{},
		&plugins.CopyEmbed{},
	}

	var kubeModules []module.Module
	if runtime.Cluster.Kubernetes.Type == common.K3s {
		kubeModules = NewK3sCreateClusterPhase(runtime)
	} else {
		kubeModules = NewCreateClusterPhase(runtime)
	}

	kubeModules = append(kubeModules,
		&gpu.InstallDepsModule{Skip: !runtime.Arg.GPU.Enable},
		&gpu.RestartK3sServiceModule{Skip: !runtime.Arg.GPU.Enable},
		&gpu.RestartContainerdModule{Skip: !runtime.Arg.GPU.Enable},
		&gpu.InstallPluginModule{Skip: !runtime.Arg.GPU.Enable},
	)

	m = append(m, kubeModules...)

	return &pipeline.Pipeline{
		Name:    "Install Terminus",
		Modules: m,
		Runtime: runtime,
	}
}
