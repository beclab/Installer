package system

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/gpu"
	_ "bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrepareSystemPhase(runtime *common.KubeRuntime) *pipeline.Pipeline {
	var isK3s = strings.Contains(runtime.Arg.KubernetesVersion, "k3s")
	var osSupport = isSupportOs()
	manifestMap, err := manifest.ReadAll(runtime.Arg.Manifest)
	if err != nil {
		logger.Fatal(err)
	}

	(&gpu.CheckWslGPU{}).Execute(runtime)

	m := []module.Module{
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		&precheck.PreCheckOsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
		},
		&patch.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
		},
		&os.ConfigSystemModule{},

		// unitl now, system ready
		&storage.InitStorageModule{Skip: runtime.Arg.WSL || !runtime.Arg.IsCloudInstance},
		&storage.InstallMinioModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: runtime.Arg.WSL || runtime.Arg.Storage.StorageType != common.Minio,
		},
		&storage.InstallRedisModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: runtime.Arg.WSL,
		},
		&storage.InstallJuiceFsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: runtime.Arg.WSL,
		},

		&container.InstallContainerModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip:        isK3s || runtime.Arg.IsCloudInstance,
			NoneCluster: true,
		}, //
		&k3s.InstallContainerModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: !isK3s || runtime.Arg.IsCloudInstance,
		},
		&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: runtime.Arg.SkipPullImages || runtime.Arg.IsCloudInstance,
		}, //

		// GPU install if necessary
		&gpu.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  runtime.Arg.BaseDir,
			},
			Skip: !runtime.Arg.GPU.Enable || !osSupport,
		},
		&gpu.RestartContainerdModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},

		// mark system prepared well
		&terminus.PreparedModule{BaseDir: runtime.Arg.BaseDir},
	}

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
