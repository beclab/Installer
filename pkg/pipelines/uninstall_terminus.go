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
	kubeVersion, _ := phase.GetKubeVersion()
	kubeType := "k8s"
	if strings.Contains(kubeVersion, "k3s") {
		kubeType = "k3s"
	}
	// var deleteCache, err = formatDeleteCache(opt)
	// if err != nil {
	// 	return err
	// }

	var arg = common.NewArgument()
	arg.SetTerminusVersion(opt.Version)
	arg.SetBaseDir(opt.BaseDir)
	arg.SetKubeVersion(kubeVersion, kubeType)
	arg.SetMinikube(opt.MiniKube, "")
	// arg.SetDeleteCache(deleteCache)
	arg.SetDeleteCRI(opt.All || (opt.Phase == cluster.PhasePrepare.String() || opt.Phase == cluster.PhaseDownload.String()))
	arg.SetStorage(&common.Storage{
		StorageVendor: os.Getenv(common.EnvCloudInstanceName),
		StorageType:   os.Getenv(common.EnvStorageTypeName),
		StorageBucket: os.Getenv(common.EnvStorageBucketName),
	})

	if err := checkPhase(opt.Phase, opt.All); err != nil {
		return err
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	phaseName := opt.Phase
	if opt.All {
		phaseName = cluster.PhaseDownload.String()
	}

	var p = cluster.UninstallTerminus(phaseName, arg, runtime)
	if err := p.Start(); err != nil {
		logger.Errorf("uninstall terminus failed: %v", err)
		return err
	}

	return nil

}

func checkPhase(phase string, all bool) error {
	if constants.OsType == common.Linux && !all {
		if cluster.UninstallPhaseString(phase).Type() == cluster.PhaseInvalid {
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

func formatDeleteCache(opt *options.CliTerminusUninstallOptions) (bool, error) {
	var all = opt.All
	var minikube = opt.MiniKube
	var quiet = opt.Quiet
	if minikube && opt.Phase == cluster.PhaseInstall.String() {
		opt.Phase = cluster.PhasePrepare.String()
	}

	if all {
		opt.Phase = cluster.PhaseDownload.String()
	}

	var phase = opt.Phase
	if phase != cluster.PhaseDownload.String() {
		return false, nil
	}

	var deleteCache = (all || phase == cluster.PhaseDownload.String())
	var input string
	var err error
	if !quiet {
		if deleteCache && !all {
			input, err = readDeleteCacheInput()
			if err != nil {
				return false, err
			}
		} else {
			input = common.YES
		}
		return strings.EqualFold(input, common.YES), nil
	} else {
		return deleteCache, nil
	}
}
