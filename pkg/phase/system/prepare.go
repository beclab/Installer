package system

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/binaries"
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	_ "bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrepareSystemPhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	var isK3s = strings.Contains(runtime.Arg.KubernetesVersion, "k3s")
	var osSupport = isSupportOs()
	m := []module.Module{
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		&terminus.TidyPackageModule{},
		// &terminus.InstallWizardDownloadModule{Version: runtime.Arg.TerminusVersion},
		&storage.DownloadStorageBinariesModule{},
		//
		&precheck.PreCheckOsModule{},
		&patch.InstallDepsModule{},
		&os.ConfigSystemModule{},
		//
		// &storage.InitStorageModule{Skip: !runtime.Arg.IsCloudInstance},
		// &storage.InstallMinioModule{Skip: runtime.Arg.Storage.StorageType != common.Minio},
		// &storage.InstallRedisModule{},
		// &binaries.K3sNodeBinariesModule{},
		//
		&binaries.NodeBinariesModule{},
		&container.InstallContainerModule{Skip: isK3s, NoneCluster: true}, //
		// &k3s.InstallContainerModule{Skip: !isK3s},
		&images.PreloadImagesModule{Skip: runtime.Arg.SkipPullImages}, //
		// &terminus.CopyToWizardModule{},
	}

	m = append(m,
		&gpu.InstallDepsModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},
		&gpu.RestartK3sServiceModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},
		&gpu.RestartContainerdModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},
		&gpu.InstallPluginModule{Skip: true},
		&terminus.PreparedModule{},
	)

	return &pipeline.Pipeline{
		Name:    "Prepare the System Environment",
		Modules: m,
		Runtime: runtime,
	}
}

func isSupportOs() bool {
	if constants.OsPlatform == common.Ubuntu && (strings.Contains(constants.OsVersion, "20.") || strings.Contains(constants.OsVersion, "22.")) {
		return true
	}

	return false
}
