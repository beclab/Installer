package storage

import (
	"fmt"
	"os/exec"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"github.com/pkg/errors"
)

// - InstallMinioClusterModule
type InstallMinioClusterModule struct {
	common.KubeModule
}

func (m *InstallMinioClusterModule) Init() {
	m.Name = "InstallMinioCluster"
}

// ~ InstallMinioOperator
type InstallMinioOperator struct {
	common.KubeAction
}

func (t *InstallMinioOperator) Execute(runtime connector.Runtime) error {
	var arch = constants.OsArch
	binary := files.NewKubeBinary("minio-operator", arch, kubekeyapiv1alpha2.DefaultMinioOperatorVersion, runtime.GetWorkDir())

	if err := binary.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
	}

	var exists = util.IsExist(binary.Path())
	if exists {
		p := binary.Path()
		if err := binary.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		} else {
			return nil
		}
	}

	if !exists || binary.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)
		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("tar zxvf %s", binary.Path()), false, true); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("install -m 755 %s/minio-operator %s", binary.BaseDir, MinioOperatorFile), false, true); err != nil {
		return err
	}

	var minioData, _ = t.PipelineCache.GetMustString(common.CacheMinioDataPath)
	var minioPassword, _ = t.PipelineCache.GetMustString(common.CacheMinioPassword)
	var cmd = fmt.Sprintf("%s init --address %s --cafile /etc/ssl/etcd/ssl/ca.pem --certfile /etc/ssl/etcd/ssl/node-%s.pem --keyfile /etc/ssl/etcd/ssl/node-%s-key.pem --volume %s --password %s",
		MinioOperatorFile, constants.LocalIp, runtime.RemoteHost().GetName(),
		runtime.RemoteHost().GetName(), minioData, minioPassword)

	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}

	return nil
}
