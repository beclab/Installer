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
	// if !opts.MiniKube {
	// 	if kubeVersion := phase.GetCurrentKubeVersion(); kubeVersion != "" {
	// 		return fmt.Errorf("Kubernetes %s is already installed. You need to uninstall it before reinstalling.", kubeVersion)
	// 	}
	// } else {
	// 	if err := checkMacOSParams(opts.MiniKube, opts.MiniKubeProfile); err != nil {
	// 		return err
	// 	}
	// }

	var terminusVersion, _ = phase.GetTerminusVersion()
	if terminusVersion != "" {
		fmt.Printf("Terminus is already installed, please uninstall it first.")
		return nil
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(getKubeVersion(opts.KubeType), opts.KubeType)
	arg.SetTerminusVersion(opts.Version)
	arg.SetMinikube(opts.MiniKube, opts.MiniKubeProfile)
	arg.SetTokenMaxAge()

	if err := arg.ArgValidate(); err != nil { // todo validate gpu for platform and os version
		return err
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return nil
	}

	manifest := opts.Manifest
	home := runtime.GetHomeDir()
	if manifest == "" {
		manifest = home + "/.terminus/installation.manifest"
	}

	runtime.Arg.SetManifest(manifest)

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

func getKubeVersion(kubeType string) string {
	if kubeType == common.K3s {
		return common.DefaultK3sVersion
	}
	return common.DefaultK8sVersion
}
