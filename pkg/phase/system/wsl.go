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
	"bytetrade.io/web3os/installer/pkg/terminus"
)

var _ phaseBuilder = &linuxPhaseBuilder{}

type wslPhaseBuilder struct {
	runtime     *common.KubeRuntime
	manifestMap manifest.InstallationManifest
}

func (l *wslPhaseBuilder) base() phase {
	return []module.Module{
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
}

func (l *wslPhaseBuilder) installContainerModule() []module.Module {
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

func (l *wslPhaseBuilder) build() []module.Module {
	(&gpu.CheckWslGPU{}).Execute(l.runtime)
	return l.base().
		addModule(l.installContainerModule()...).
		addModule(&images.PreloadImagesModule{
			ManifestModule: manifest.ManifestModule{
				Manifest: l.manifestMap,
				BaseDir:  l.runtime.Arg.BaseDir,
			},
		}).
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
