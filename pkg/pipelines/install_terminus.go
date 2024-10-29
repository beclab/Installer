package pipelines

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	ctrl "bytetrade.io/web3os/installer/controllers"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func CliInstallTerminusPipeline(opts *options.CliTerminusInstallOptions) error {
	var terminusVersion, _ = phase.GetTerminusVersion()
	if terminusVersion != "" {
		fmt.Printf("Terminus is already installed, please uninstall it first.")
		return nil
	}

	arg := common.NewArgument()
	arg.SetBaseDir(opts.BaseDir)
	arg.SetKubeVersion(opts.KubeType)
	arg.SetTerminusVersion(opts.Version)
	arg.SetMinikubeProfile(opts.MiniKubeProfile)
	arg.SetStorage(getStorageValueFromEnv())
	arg.SetReverseProxy()
	arg.SetTokenMaxAge()

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		fmt.Printf("Error creating installation runtime: %v\n", err)
		return nil
	}

	manifest := opts.Manifest
	if manifest == "" {
		manifest = path.Join(runtime.GetInstallerDir(), "installation.manifest")
	}

	runtime.Arg.SetManifest(manifest)

	// var p = cluster.CreateTerminus(runtime)
	var p = cluster.InstallSystemPhase(runtime)
	if err := p.Start(); err != nil {
		return fmt.Errorf("create terminus error %v", err)
	}

	if !runtime.GetSystemInfo().IsWindows() {
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
	}

	return nil
}
