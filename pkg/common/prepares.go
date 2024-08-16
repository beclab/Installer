package common

import (
	"fmt"
	"path"
	"strconv"

	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"github.com/pkg/errors"
)

type Skip struct {
	KubePrepare
	Not bool
}

func (p *Skip) PreCheck(runtime connector.Runtime) (bool, error) {
	return !p.Not, nil
}

type Stop struct {
	prepare.BasePrepare
}

func (p *Stop) PreCheck(runtime connector.Runtime) (bool, error) {
	return true, nil
	// return false, fmt.Errorf("STOP !!!!!!")
}

type GetCommandKubectl struct {
	prepare.BasePrepare
}

func (p *GetCommandKubectl) PreCheck(runtime connector.Runtime) (bool, error) {

	cmd, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("command -v %s", CommandKubectl), false, false)
	if err != nil {
		return true, nil
	}
	if cmd != "" {
		p.PipelineCache.Set(CacheKubectlKey, cmd)
	}
	return true, nil
}

type GetKubeVersion struct {
	prepare.BasePrepare
}

func (p *GetKubeVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	return true, nil
}

type GetMasterNum struct {
	prepare.BasePrepare
}

func (p *GetMasterNum) PreCheck(runtime connector.Runtime) (bool, error) {
	var kubectlpath, _ = p.PipelineCache.GetMustString(CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath = path.Join(BinDir, CommandKubectl)
	}

	var cmd = fmt.Sprintf("%s get node | awk '{if(NR>1){print $3}}' | grep master | wc -l", kubectlpath)
	var stdout, err = runtime.GetRunner().SudoCmd(cmd, false, false)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "get master num failed")
	}

	masterNum, _ := strconv.ParseInt(stdout, 10, 64)

	p.PipelineCache.Set(CacheMasterNum, masterNum)

	return true, nil
}

type GetNodeNum struct {
	prepare.BasePrepare
}

func (p *GetNodeNum) PreCheck(runtime connector.Runtime) (bool, error) {
	var kubectlpath, _ = p.PipelineCache.GetMustString(CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath = path.Join(BinDir, CommandKubectl)
	}

	var cmd = fmt.Sprintf("%s get node | wc -l", kubectlpath)
	var stdout, err = runtime.GetRunner().SudoCmd(cmd, false, false)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "get node num failed")
	}

	nodeNum, _ := strconv.ParseInt(stdout, 10, 64)

	p.PipelineCache.Set(CacheNodeNum, nodeNum)

	return true, nil
}

type ClusterType struct {
	KubePrepare
	ClusterType string
	Not         bool
}

func (p *ClusterType) PreCheck(runtime connector.Runtime) (bool, error) {
	if p.KubeConf == nil || p.KubeConf.Cluster == nil {
		return false, nil
	}
	var isK3s = p.KubeConf.Cluster.Kubernetes.Type == p.ClusterType
	if p.Not {
		return !isK3s, nil
	}

	return isK3s, nil
}

type OsType struct {
	KubePrepare
	OsType string
	Not    bool
}

func (p *OsType) PreCheck(runtime connector.Runtime) (bool, error) {
	isOs := constants.OsType == p.OsType
	if p.Not {
		return !isOs, nil
	}
	return isOs, nil
}

type OsVersion struct {
	KubePrepare
	OsVersion map[string]bool
}

func (p *OsVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	if p.OsVersion == nil || len(p.OsVersion) == 0 {
		return false, nil
	}

	var flag = false
	for k, v := range p.OsVersion {
		if k == constants.OsVersion {
			flag = v
			break
		}
	}

	return flag, nil
}
