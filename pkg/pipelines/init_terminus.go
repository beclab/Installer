package pipelines

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"

	ctrl "bytetrade.io/web3os/installer/controllers"
	"bytetrade.io/web3os/installer/pkg/bootstrap/os"
	"bytetrade.io/web3os/installer/pkg/bootstrap/patch"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func InstallTerminusPipeline(args common.Argument) error {
	runtime, err := common.NewKubeRuntime(common.AllInOne, args) //
	if err != nil {
		return err
	}

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.PreCheckOsModule{}, // precheck_os()
		&patch.InstallDepsModule{},   // install_deps
		&os.ConfigSystemModule{},     // config_system
		// &images.PreloadImagesModule{},
	}

	var kubeModules []module.Module
	if runtime.Cluster.Kubernetes.Type == common.K3s {
		// FIXME:
		kubeModules = cluster.NewK3sCreateClusterPhase(runtime, nil) // +
	} else {
		kubeModules = cluster.NewCreateClusterPhase(runtime, nil)
	}

	m = append(m, kubeModules...)

	p := pipeline.Pipeline{
		Name:    "Install Terminus",
		Modules: m,
		Runtime: runtime,
	}

	go func() {
		if err := p.Start(); err != nil {
			logger.Errorf("install terminus failed %v", err)
			return
		}

		if runtime.Arg.InCluster {
			if err := ctrl.UpdateStatus(runtime); err != nil {
				logger.Errorf("failed to update status: %v", err)
				return
			}
			kkConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.ObjName))
			if config, err := ioutil.ReadFile(kkConfigPath); err != nil {
				logger.Errorf("failed to read kubeconfig: %v", err)
				return
			} else {
				runtime.Kubeconfig = base64.StdEncoding.EncodeToString(config)
				if err := ctrl.UpdateKubeSphereCluster(runtime); err != nil {
					logger.Errorf("failed to update kubesphere cluster: %v", err)
					return
				}
				if err := ctrl.SaveKubeConfig(runtime); err != nil {
					logger.Errorf("failed to save kubeconfig: %v", err)
					return
				}
			}
		}
	}()

	return nil
}
