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
	"bytetrade.io/web3os/installer/pkg/terminus"
)

func PrepareSystemPhase(runtime *common.KubeRuntime, manifestPath, baseDir string) *pipeline.Pipeline {
	var isK3s = strings.Contains(runtime.Arg.KubernetesVersion, "k3s")
	var osSupport = isSupportOs()
	manifestMap, err := manifest.ReadAll(manifestPath)
	if err != nil {
		logger.Fatal(err)
	}

	m := []module.Module{
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		// &terminus.TidyPackageModule{},
		// &storage.DownloadStorageBinariesModule{},
		//
		&precheck.PreCheckOsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
		},
		&patch.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
		},
		&os.ConfigSystemModule{},
		// &binaries.K3sNodeBinariesModule{},
		// &binaries.NodeBinariesModule{},
		&container.InstallContainerModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
			Skip:        isK3s,
			NoneCluster: true,
		}, //
		&k3s.InstallContainerModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
			Skip: !isK3s,
		},
		&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
			Skip: runtime.Arg.SkipPullImages,
		}, //
		// &terminus.CopyToWizardModule{},
		&gpu.CheckWSLGPUEnableModule{},
		&gpu.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: manifestMap,
				BaseDir:  baseDir,
			},
			Skip: !runtime.Arg.GPU.Enable || !osSupport,
		},
		// &gpu.RestartK3sServiceModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},
		&gpu.RestartContainerdModule{Skip: !runtime.Arg.GPU.Enable || !osSupport},
		// &gpu.InstallPluginModule{Skip: true},
		&terminus.PreparedModule{},
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
