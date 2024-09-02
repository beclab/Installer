package system

import (
	"strings"

	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/gpu"
	"bytetrade.io/web3os/installer/pkg/images"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/storage"
	"bytetrade.io/web3os/installer/pkg/terminus"
)

var _ phaseBuilder = &linuxPhaseBuilder{}

type linuxPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (l *linuxPhaseBuilder) base() phase {
	m := []module.Module{
		&precheck.GetSysInfoModel{},
		&plugins.CopyEmbed{},
		&precheck.PreCheckOsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
		},
		&patch.InstallDepsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
		},
		&os.ConfigSystemModule{},
	}

	return m
}

func (l *linuxPhaseBuilder) storage() phase {
	return []module.Module{
		&storage.InstallMinioModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
			Skip: l.runtime.Arg.Storage.StorageType != common.Minio,
		},
		&storage.InstallRedisModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
		},
		&storage.InstallJuiceFsModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
		},
	}
}

func (l *linuxPhaseBuilder) installContainerModule() []module.Module {
	var isK3s = strings.Contains(l.runtime.Arg.KubernetesVersion, "k3s")
	if isK3s {
		return []module.Module{
			&k3s.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.runtime.Arg.BaseDir,
				},
			},
		}
	} else {
		return []module.Module{
			&container.InstallContainerModule{
				ManifestModule: manifest.ManifestModule{
					Manifest: l.manifestMap,
					BaseDir:  l.runtime.Arg.BaseDir,
				},
				NoneCluster: true,
			}, //
		}
	}
}

func (l *linuxPhaseBuilder) build() []module.Module {
	return l.base().
		addModule(cloudModuleBuilder(func() []module.Module {
			return []module.Module{
				&storage.InitStorageModule{Skip: !l.runtime.Arg.IsCloudInstance},
			}
		}).withCloud(l.runtime)...).
		addModule(l.storage()...).
		addModule(cloudModuleBuilder(l.installContainerModule).withoutCloud(l.runtime)...).
		addModule(cloudModuleBuilder(func() []module.Module {
			// unitl now, system ready
			return []module.Module{
				&images.PreloadImagesModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.Arg.BaseDir,
					},
				}, //
			}
		}).withoutCloud(l.runtime)...).
		addModule(gpuModuleBuilder(func() []module.Module {
			return []module.Module{
				&gpu.InstallDepsModule{
					ManifestModule: manifest.ManifestModule{
						Manifest: l.manifestMap,
						BaseDir:  l.runtime.Arg.BaseDir,
					},
				},
				&gpu.RestartContainerdModule{},
			}

		}).withGPU(l.runtime)...).
		addModule(&terminus.PreparedModule{BaseDir: l.runtime.Arg.BaseDir})
}
