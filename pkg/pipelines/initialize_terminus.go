package pipelines

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	ctrl "bytetrade.io/web3os/installer/controllers"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func CliInitializeTerminusPipeline(kubeType string, minikube bool, minikubeProfileName, registryMirrors string) error {
	if err := checkMacOSParams(minikube, minikubeProfileName); err != nil {
		return err
	}

	// var ksVersion, err = getNodeVersion(kubeType, minikube)
	// if err != nil {
	// 	return err
	// }

	var arg = common.NewArgument()
	arg.RegistryMirrors = registryMirrors
	// arg.SetKubernetesVersion(kubeType, ksVersion)
	arg.SetMinikube(minikube, minikubeProfileName)

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var p = cluster.InitKube(*arg, runtime)
	if err := p.Start(); err != nil {
		return err
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

func checkMacOSParams(minikube bool, minikubeProfileName string) error {
	if constants.OsPlatform == common.Darwin && !minikube {
		return fmt.Errorf("MacOS startup parameter error, need to specify parameters --minikube --profile PROFILE_NAME")
	}

	if minikube && len(minikubeProfileName) == 0 {
		return fmt.Errorf("minikube profile name cannot be empty")
	}
	return nil
}

func getNodeVersion(kubeType string, minikube bool) (string, error) {
	var ver string
	if !minikube {
		switch kubeType {
		case common.K8s:
			ver = common.DefaultK8sVersion
		case common.K3s:
			fallthrough
		default:
			ver = common.DefaultK3sVersion
		}

		return ver, nil
	}

	if constants.OsType != common.Darwin {
		return ver, fmt.Errorf("Start minikube, but the system type is incorrect, not a Darwin system.")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var cmd = fmt.Sprintf("/usr/local/bin/kubectl get nodes -o jsonpath='{.items[0].status.nodeInfo.kubeletVersion}'")
	stdout, _, err := util.ExecWithContext(ctx, cmd, false, false)
	if err != nil {
		return common.DefaultK8sVersion, nil
	}

	if strings.Contains(stdout, "k3s") {
		if strings.Contains(stdout, "-") {
			stdout = strings.ReplaceAll(stdout, "-", "+")
		}

		var v1 = strings.Split(stdout, "+")
		if len(v1) != 2 {
			return ver, fmt.Errorf("parse k3s version failed %s", stdout)
		}
		ver = fmt.Sprintf("%s-k3s", v1[0])
	} else {
		ver = stdout
	}

	return ver, nil
}
