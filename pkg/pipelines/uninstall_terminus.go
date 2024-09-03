package pipelines

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func UninstallTerminusPipeline(opt *options.CliTerminusUninstallOptions) error {
	var kubeVersion = phase.GetCurrentKubeVersion()
	var deleteCache, err = formatDeleteCache(opt.Quiet, opt.DeleteCache, opt.MiniKube, opt.Phase)
	if err != nil {
		return err
	}

	deleteCache = false
	var arg = common.NewArgument()
	arg.SetKubernetesVersion(kubeVersion, kubeVersion)
	arg.SetMinikube(opt.MiniKube, "")
	arg.SetDeleteCache(deleteCache)
	arg.SetStorage(&common.Storage{
		StorageType:   formatParms(common.EnvStorageTypeName, opt.StorageType),
		StorageBucket: formatParms(common.EnvStorageBucketName, opt.StorageBucket),
	})

	phaseName := opt.Phase
	if err := checkPhase(phaseName); err != nil {
		return err
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	home := runtime.GetHomeDir()
	baseDir := opt.BaseDir
	if baseDir == "" {
		baseDir = home + "/.terminus"
	}

	var p = cluster.UninstallTerminus(baseDir, phaseName, arg, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("uninstall terminus failed: %v", err)
		return err
	}

	return nil

}

func checkPhase(phase string) error {
	if constants.OsType == common.Linux {
		if phase == "" || (phase != "install" && phase != "prepare" && phase != "download") {
			return fmt.Errorf("Please specify the phase to uninstall, such as --phase install. Supported: install, prepare, download.")
		}
	}
	return nil
}

func readDeleteCacheInput() (string, error) {
	reader := bufio.NewReader(os.Stdin)

LOOP:
	fmt.Printf("\nDelete the locally stored image files? The installation system will prioritize loading local image files. [yes/no]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input != common.YES && input != common.NO {
		goto LOOP
	}

	return input, nil
}

func formatParms(key, val string) string {
	valEnv := os.Getenv(key)
	if !strings.EqualFold(valEnv, "") {
		return valEnv
	}
	if !strings.EqualFold(val, "") {
		return val
	}
	return ""
}

func formatDeleteCache(quiet, deleteCache, minikube bool, phase string) (bool, error) {
	if !minikube && phase != "download" {
		return false, nil
	}

	var input string
	var err error
	if !quiet {
		if !deleteCache {
			input, err = readDeleteCacheInput()
			if err != nil {
				return false, err
			}
		} else {
			input = "true"
		}
	} else {
		if deleteCache {
			input = "true"
		}
	}

	return strings.EqualFold(input, common.TRUE), nil
}
