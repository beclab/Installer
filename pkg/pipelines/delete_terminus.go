package pipelines

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
	"bytetrade.io/web3os/installer/pkg/storage"
)

func UninstallTerminusPipeline(opt *options.CliTerminusUninstallOptions) error {
	var kubeVersion = phase.GetCurrentKubeVersion()
	var deleteCache, err = formatDeleteCache(opt.Quiet, opt.DeleteCache)
	if err != nil {
		return err
	}

	var arg = common.NewArgument()
	arg.SetKubernetesVersion(kubeVersion, kubeVersion)
	arg.SetMinikube(opt.MiniKube, "")
	arg.SetDeleteCache(deleteCache)
	arg.SetStorage(&common.Storage{
		StorageType:   formatParms(common.EnvStorageTypeName, opt.StorageType),
		StorageBucket: formatParms(common.EnvStorageBucketName, opt.StorageBucket),
	})

	runtime, err := common.NewKubeRuntime(common.AllInOne, *arg)
	if err != nil {
		return err
	}

	var m = []module.Module{&precheck.GreetingsModule{}, &precheck.GetSysInfoModel{}}

	switch constants.OsType {
	case common.Darwin:
		m = append(m, cluster.DeleteMinikubePhase(*arg, runtime)...)
	default:
		m = append(m, &precheck.GetStorageKeyModule{}, &storage.RemoveMountModule{})
		m = append(m, cluster.DeleteClusterPhase(runtime)...)
	}

	p := pipeline.Pipeline{
		Name:    "Delete Terminus",
		Runtime: runtime,
		Modules: m,
	}

	if err := p.Start(); err != nil {
		logger.Errorf("delete terminus failed: %v", err)
		return err
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

func formatDeleteCache(quiet, deleteCache bool) (bool, error) {
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
