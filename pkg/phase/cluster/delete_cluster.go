package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/filesystem"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/storage"
)

type UninstallPhaseType int

const (
	PhaseInvalid UninstallPhaseType = iota
	PhaseInstall
	PhasePrepare
	PhaseDownload
)

type phaseBuilder struct {
	minikube bool
	phase    string
	baseDir  string
	modules  []module.Module
	runtime  *common.KubeRuntime
}

func (p *phaseBuilder) convert() UninstallPhaseType {
	switch p.phase {
	case "install":
		return PhaseInstall
	case "prepare":
		return PhasePrepare
	case "download":
		return PhaseDownload
	}
	return PhaseInvalid
}

func (p *phaseBuilder) fin() *phaseBuilder {
	if p.minikube {
		p.modules = append([]module.Module{
			&kubesphere.DeleteCacheModule{},
			&kubesphere.DeleteMinikubeModule{},
			&filesystem.DeleteInstalledModule{},
		}, p.modules...)
	} else {
		p.modules = append(
			[]module.Module{
				&precheck.GetStorageKeyModule{},
				&storage.RemoveMountModule{},
			},
			p.modules...)
	}

	p.modules = append(
		[]module.Module{
			&precheck.GreetingsModule{},
			&precheck.GetSysInfoModel{},
		},
		p.modules...)

	return p
}

func (p *phaseBuilder) phaseInstall() *phaseBuilder {
	if p.minikube {
		return p
	}

	if p.convert() >= PhaseInstall {
		p.modules = append(p.modules,
			&kubernetes.ResetClusterModule{},
			&k3s.DeleteClusterModule{},
			&os.ClearOSEnvironmentModule{},
			&certs.UninstallAutoRenewCertsModule{},
			&container.KillContainerdProcessModule{},
			&k3s.UninstallK3sModule{},
		)
	}
	return p
}

func (p *phaseBuilder) phasePrepare() *phaseBuilder {
	if p.minikube {
		return p
	}

	if p.convert() >= PhasePrepare {
		p.modules = append(p.modules, []module.Module{
			&container.DeleteZfsMountModule{},
			&storage.RemoveStorageModule{},
			&container.UninstallContainerModule{},
		}...)
	}
	return p
}

func (p *phaseBuilder) phaseDownload() *phaseBuilder {
	if p.minikube {
		return p
	}

	if p.convert() >= PhaseDownload {
		p.modules = append(p.modules, []module.Module{
			&filesystem.DeleteInstalledModule{
				BaseDir: p.baseDir,
			},
		}...)
	}

	return p
}

func UninstallTerminus(baseDir, phase string, args *common.Argument, runtime *common.KubeRuntime) pipeline.Pipeline {
	var builder = &phaseBuilder{
		minikube: args.Minikube,
		phase:    phase,
		baseDir:  baseDir,
		runtime:  runtime,
	}
	builder.phaseInstall().phasePrepare().phaseDownload().fin()
	return pipeline.Pipeline{
		Name:    "Uninstall Terminus",
		Runtime: builder.runtime,
		Modules: builder.modules,
	}
}
