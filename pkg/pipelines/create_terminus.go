package pipelines

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	ctrl "bytetrade.io/web3os/installer/controllers"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func CliInstallTerminusPipeline(opts *options.CliTerminusInstallOptions) error {
	if kubeVersion := phase.GetCurrentKubeVersion(); kubeVersion != "" {
		return fmt.Errorf("Kubernetes %s is already installed. You need to uninstall it before reinstalling.", kubeVersion)
	}

	var userParms = phase.UserParameters()
	// var storageParms = phase.StorageParameters()

	arg := common.NewArgument()
	arg.SetKubernetesVersion(opts.KubeType, "")
	// arg.SetStorage(storageParms) // todo
	arg.SetMinikube(opts.MiniKube, opts.MiniKubeProfile)
	arg.SetProxy(opts.Proxy, opts.RegistryMirrors)

	arg.User = userParms

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return nil
	}

	var p = cluster.CreateTerminus(*arg, runtime)
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
