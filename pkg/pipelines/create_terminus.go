package pipelines

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	ctrl "bytetrade.io/web3os/installer/controllers"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func CliInstallTerminusPipeline(kubeType string, proxy string) error {
	if kubeVersion := phase.GetCurrentKubeVersion(); kubeVersion != "" {
		return fmt.Errorf("Kubernetes %s already installed", kubeVersion)
	}

	var userParms = phase.UserParameters()
	var storageParms = phase.StorageParameters()
	var sp, _ = json.Marshal(storageParms)
	fmt.Printf("STORAGE: %s\n", string(sp))

	arg := common.Argument{
		KsEnable:         true,
		KsVersion:        common.DefaultKubeSphereVersion,
		InstallPackages:  false,
		SKipPushImages:   false,
		ContainerManager: common.Containerd,
		User:             userParms,
		Storage:          storageParms,
	}

	if proxy != "" {
		arg.Proxy = proxy
	}

	switch kubeType {
	case common.K3s:
		arg.KubernetesVersion = common.DefaultK3sVersion
	case common.K8s:
		arg.KubernetesVersion = common.DefaultK8sVersion
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, arg)
	if err != nil {
		return nil
	}

	var p = cluster.CreateTerminus(arg, runtime)
	if err := p.Start(); err != nil {
		return fmt.Errorf("create terminus error %v", err)
	}

	if runtime.Arg.InCluster {
		if err := ctrl.UpdateStatus(runtime); err != nil {
			logger.Errorf("failed to update status: %v", err)
			return err
		}
		kkConfigPath := filepath.Join(runtime.GetWorkDir(), fmt.Sprintf("config-%s", runtime.ObjName))
		if config, err := ioutil.ReadFile(kkConfigPath); err != nil {
			logger.Errorf("failed to read kubeconfig: %v", err)
			return err
		} else {
			runtime.Kubeconfig = base64.StdEncoding.EncodeToString(config)
			if err := ctrl.UpdateKubeSphereCluster(runtime); err != nil {
				logger.Errorf("failed to update kubesphere cluster: %v", err)
				return err
			}
			if err := ctrl.SaveKubeConfig(runtime); err != nil {
				logger.Errorf("failed to save kubeconfig: %v", err)
				return err
			}
		}
	}

	return nil
}
