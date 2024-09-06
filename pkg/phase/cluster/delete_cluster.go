package cluster

import (
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/certs"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/container"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/k3s"
	"bytetrade.io/web3os/installer/pkg/kubernetes"
	"bytetrade.io/web3os/installer/pkg/kubesphere"
	"bytetrade.io/web3os/installer/pkg/storage"
)

type UninstallPhaseType int
type UninstallPhaseString string

const (
	PhaseInvalid UninstallPhaseType = iota
	PhaseInstall
	PhasePrepare
	PhaseDownload
)

func (p UninstallPhaseType) String() string {
	switch p {
	case PhaseInvalid:
		return "invalid"
	case PhaseInstall:
		return "install"
	case PhasePrepare:
		return "prepare"
	case PhaseDownload:
		return "download"
	}
	return ""
}

func (s UninstallPhaseString) String() string {
	return string(s)
}

func (s UninstallPhaseString) Type() UninstallPhaseType {
	switch s.String() {
	case PhaseInstall.String():
		return PhaseInstall
	case PhasePrepare.String():
		return PhasePrepare
	case PhaseDownload.String():
		return PhaseDownload
	}
	return PhaseInvalid

}

type phaseBuilder struct {
	phase   string
	modules []module.Module
	runtime *common.KubeRuntime
}

func (p *phaseBuilder) convert() UninstallPhaseType {
	return UninstallPhaseString(p.phase).Type()
}

func (p *phaseBuilder) phaseInstall() *phaseBuilder {
	if p.convert() >= PhaseInstall {
		_ = (&kubesphere.GetKubeType{}).Execute(p.runtime)

		p.modules = []module.Module{
			&precheck.GreetingsModule{},
			&precheck.GetSysInfoModel{},
			&precheck.GetStorageKeyModule{},
			&storage.RemoveMountModule{},
		}

		if p.runtime.Arg.Storage.StorageType == common.S3 || p.runtime.Arg.Storage.StorageType == common.OSS {
			p.modules = append(p.modules,
				&precheck.GetStorageKeyModule{},
				&storage.RemoveMountModule{},
			)
		}

		switch p.runtime.Cluster.Kubernetes.Type {
		case common.K3s:
			p.modules = append(p.modules, &k3s.DeleteClusterModule{})
		default:
			p.modules = append(p.modules, &kubernetes.ResetClusterModule{}, &kubernetes.UmountKubeModule{})
		}

		p.modules = append(p.modules,
			&os.ClearOSEnvironmentModule{},
			&certs.UninstallAutoRenewCertsModule{},
			&container.KillContainerdProcessModule{},
			&storage.DeleteUserDataModule{},
			&storage.DeletePhaseFlagModule{
				PhaseFile: ".installed",
				BaseDir:   p.runtime.GetBaseDir(),
			},
		)
	}
	return p
}

func (p *phaseBuilder) phasePrepare() *phaseBuilder {
	if p.convert() >= PhasePrepare {
		p.modules = append(p.modules,
			&container.DeleteZfsMountModule{},
			&storage.RemoveStorageModule{},
			&container.UninstallContainerModule{},
			&storage.DeleteTerminusDataModule{},
			&storage.DeletePhaseFlagModule{
				PhaseFile: ".prepared",
				BaseDir:   p.runtime.GetBaseDir(),
			},
		)
	}
	return p
}

func (p *phaseBuilder) phaseDownload() *phaseBuilder {
	if p.convert() >= PhaseDownload && p.runtime.Arg.DeleteCache {
		p.modules = append(p.modules, &storage.DeleteCacheModule{
			BaseDir: p.runtime.GetBaseDir(),
		})
	}

	return p
}

func (p *phaseBuilder) phaseMacos() {
	p.modules = []module.Module{
		&precheck.GreetingsModule{},
		&precheck.GetSysInfoModel{},
		&kubesphere.DeleteCacheModule{},
		&kubesphere.DeleteMinikubeModule{},
	}
}

func UninstallTerminus(phase string, args *common.Argument, runtime *common.KubeRuntime) pipeline.Pipeline {
	var builder = &phaseBuilder{
		phase:   phase,
		runtime: runtime,
	}

	if args.Minikube {
		builder.phaseMacos()
	} else {
		builder.
			phaseInstall().
			phasePrepare().
			phaseDownload()
	}

	return pipeline.Pipeline{
		Name:    "Uninstall Terminus",
		Runtime: builder.runtime,
		Modules: builder.modules,
	}
}
