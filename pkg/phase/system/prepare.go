package system

import (
	"bytetrade.io/web3os/installer/pkg/binaries"
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrepareSystemPhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		&terminus.InstallWizardDownloadModule{Version: runtime.Arg.TerminusVersion},
		&precheck.PreCheckOsModule{},
		&patch.InstallDepsModule{},
		&os.ConfigSystemModule{},
		&storage.DownloadStorageBinariesModule{},
		&storage.InitStorageModule{Skip: !runtime.Arg.IsCloudInstance},
		// &storage.InstallMinioModule{Skip: runtime.Arg.Storage.StorageType != common.Minio},
		// &storage.InstallRedisModule{},
		&binaries.K3sNodeBinariesModule{},
		&binaries.NodeBinariesModule{},
		&k3s.InstallContainerModule{},
		&images.PreloadImagesModule{Skip: runtime.Arg.SkipPullImages},
	}

	return &pipeline.Pipeline{
		Name:    "Prepare the System Environment",
		Modules: m,
		Runtime: runtime,
	}
}
