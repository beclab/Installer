package pipelines

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/pipeline"
	"bytetrade.io/web3os/installer/pkg/phase"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
)

func UninstallTerminusPipeline(minikube bool, deleteCache bool) error {
	var input string
	var err error
	var kubeVersion = phase.GetCurrentKubeVersion()
	var deleteCacheEnv = os.Getenv("DELETE_CACHE")

	if !deleteCache && strings.EqualFold(deleteCacheEnv, common.TRUE) {
		deleteCache = true
	}

	if !deleteCache {
		input, err = readDeleteCacheInput()
		if err != nil {
			return err
		}
	}

	var args = common.Argument{
		KubernetesVersion: kubeVersion,
		ContainerManager:  common.Containerd,
		Minikube:          minikube,
		DeleteCache:       strings.EqualFold(input, common.YES),
	}

	runtime, err := common.NewKubeRuntime(common.AllInOne, args)
	if err != nil {
		return err
	}

	var m []module.Module

	switch constants.OsType {
	case common.Darwin:
		m = append(m, cluster.DeleteMinikubePhase(args, runtime)...)
	default:
		m = append(m, &precheck.GetStorageKeyModule{})
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
	fmt.Printf("\nDelete terminus caches? [yes/no]:")
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
