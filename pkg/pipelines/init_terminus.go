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

// + 这里是正式的代码
func InstallTerminusPipeline(args common.Argument) error {
	runtime, err := common.NewKubeRuntime(common.AllInOne, args) // 后续拆解 install_cmd.sh，会用到 KubeRuntime
	if err != nil {
		return err
	}

	m := []module.Module{
		&precheck.GreetingsModule{},
		&precheck.PreCheckOsModule{}, // * 对应 precheck_os()
		&patch.InstallDepsModule{},   // * 对应 install_deps
		&os.ConfigSystemModule{},     // * 对应 config_system
		// &images.PreloadImagesModule{},
	}

	var kubeModules []module.Module
	if runtime.Cluster.Kubernetes.Type == common.K3s {
		kubeModules = cluster.NewK3sCreateClusterPhase(runtime) // + 这里开发
	} else {
		kubeModules = cluster.NewCreateClusterPhase(runtime)
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

		if runtime.Cluster.KubeSphere.Enabled {

			fmt.Print(`Installation is complete.
	
	Please check the result using the command:
	
		kubectl logs -n kubesphere-system $(kubectl get pod -n kubesphere-system -l 'app in (ks-install, ks-installer)' -o jsonpath='{.items[0].metadata.name}') -f   
	
	`)
		} else {
			fmt.Print(`Installation is complete.
	
	Please check the result using the command:
			
		kubectl get pod -A
	
	`)

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
