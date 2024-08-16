package gpu

import (
	"fmt"
	"os/exec"
	"path"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"github.com/pkg/errors"
)

type InstallCudaDeps struct {
	common.KubeAction
}

func (t *InstallCudaDeps) Execute(runtime connector.Runtime) error {
	var arch string
	switch runtime.GetRunner().Host.GetArch() {
	case common.Arm64:
		arch = "arm64"
	default:
		arch = "x86_64"
	}

	var prePath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)

	var cudakeyring = files.NewKubeBinary("cuda-keyring", arch, kubekeyapiv1alpha2.DefaultCudaKeyringVersion, prePath)

	if err := cudakeyring.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", cudakeyring.FileName)
	}

	var exists = util.IsExist(cudakeyring.Path())
	if exists {
		p := cudakeyring.Path()
		if err := cudakeyring.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		}
	}

	if !exists || cudakeyring.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, cudakeyring.ID, cudakeyring.Version)
		if err := cudakeyring.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", cudakeyring.ID, cudakeyring.Url, err)
		}
	}

	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("dpkg -i %s", cudakeyring.Path()), false, true); err != nil {
		return err
	}

	return nil
}

type InstallCudaDriver struct {
	common.KubeAction
}

func (t *InstallCudaDriver) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmdExt("apt-get update", false, true); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmdExt("apt-get -y install cuda-12-1", false, true); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmdExt("apt-get -y install nvidia-kernel-open-545", false, true); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmdExt("apt-get -y install nvidia-driver-545", false, true); err != nil {
		return err
	}

	return nil
}

type UpdateCudaSource struct {
	common.KubeAction
}

func (t *UpdateCudaSource) Execute(runtime connector.Runtime) error {
	// only for ubuntu20.04  ubunt22.04

	var version string
	if strings.Contains(constants.OsVersion, "22.") {
		version = "22.04"
	} else {
		version = "20.04"
	}
	var distribution = fmt.Sprintf("%s%s", constants.OsPlatform, version)

	var cmd = fmt.Sprintf("curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | apt-key add -")
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		return err
	}

	cmd = fmt.Sprintf("curl -s -L https://nvidia.github.io/libnvidia-container/%s/libnvidia-container.list | tee /etc/apt/sources.list.d/libnvidia-container.list", distribution)
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type InstallNvidiaContainerToolkit struct {
	common.KubeAction
}

func (t *InstallNvidiaContainerToolkit) Execute(runtime connector.Runtime) error {
	logger.Debugf("install nvidia-container-toolkit")
	if _, err := runtime.GetRunner().SudoCmdExt("apt-get update && sudo apt-get install -y nvidia-container-toolkit jq", false, true); err != nil {
		return err
	}
	return nil
}

type PatchK3s struct { // patch k3s on wsl
	common.KubeAction
}

func (t *PatchK3s) Execute(runtime connector.Runtime) error {
	return nil
}
