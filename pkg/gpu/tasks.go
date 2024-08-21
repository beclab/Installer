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
	k3sGpuTemplates "bytetrade.io/web3os/installer/pkg/gpu/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type CheckWslGPU struct {
	common.KubeAction
}

func (t *CheckWslGPU) Execute(runtime connector.Runtime) error {
	if !t.KubeConf.Arg.WSL {
		return nil
	}
	var nvidiaSmiFile = "/usr/lib/wsl/lib/nvidia-smi"
	if !util.IsExist(nvidiaSmiFile) {
		return nil
	}

	stdout, err := runtime.GetRunner().Host.CmdExt("/usr/lib/wsl/lib/nvidia-smi -L|grep 'NVIDIA'|grep UUID", false, true)
	if err != nil {
		logger.Errorf("nvidia-smi not found")
		return nil
	}
	if stdout == "" {
		return nil
	}

	t.KubeConf.Arg.SetGPU(true, true)

	return nil
}

type CopyEmbedGpuFiles struct {
	common.KubeAction
}

func (t *CopyEmbedGpuFiles) Execute(runtime connector.Runtime) error {
	var dst = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir)
	if err := utils.CopyEmbed(assets, ".", dst); err != nil {
		return err
	}

	return nil
}

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

	var cmd string
	var keyPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.GpuDir, "gpgkey")
	if !util.IsExist(keyPath) {
		cmd = fmt.Sprintf("curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | apt-key add -")
	} else {
		cmd = fmt.Sprintf("apt-key add %s", keyPath)
	}
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		return err
	}

	var libPath = path.Join(runtime.GetHomeDir(), cc.TerminusDir, cc.PackageCacheDir, cc.GpuDir, "libnvidia-container.list")
	if !util.IsExist(libPath) {
		cmd = fmt.Sprintf("curl -s -L https://nvidia.github.io/libnvidia-container/%s/libnvidia-container.list | tee /etc/apt/sources.list.d/libnvidia-container.list", distribution)
	} else {
		cmd = fmt.Sprintf("cp %s %s", libPath, "/etc/apt/sources.list.d/")
	}
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

type PatchK3sDriver struct { // patch k3s on wsl
	common.KubeAction
}

func (t *PatchK3sDriver) Execute(runtime connector.Runtime) error {
	var cmd = "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1"
	driverPath, err := runtime.GetRunner().SudoCmdExt(cmd, false, true)
	if err != nil {
		return err
	}

	if driverPath == "" {
		logger.Debugf("cuda driver not found")
		return nil
	} else {
		logger.Debugf("cuda driver found: %s", driverPath)
	}

	templateStr, err := util.Render(k3sGpuTemplates.K3sCudaFixValues, nil)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", k3sGpuTemplates.K3sCudaFixValues.Name()))
	}

	var fixName = "cuda_lib_fix.sh"
	var fixPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, "gpu", "cuda_lib_fix.sh")
	if err := util.WriteFile(fixPath, []byte(templateStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fixPath))
	}

	var dstName = path.Join(common.BinDir, fixName)
	if err := runtime.GetRunner().SudoScp(fixPath, dstName); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fixPath, dstName))
	}

	cmd = fmt.Sprintf("echo 'ExecStartPre=-/usr/local/bin/%s' >> /etc/systemd/system/k3s.service", fixName)
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmdExt("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	return nil
}

type RestartContainerd struct {
	common.KubeAction
}

func (t *RestartContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmdExt("nvidia-ctk runtime configure --runtime=containerd --set-as-default", false, true); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmdExt("systemctl restart containerd", false, true); err != nil {
		return err
	}
	return nil
}

type InstallPlugin struct {
	common.KubeAction
}

func (t *InstallPlugin) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	var pluginFile = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir, "nvidia-device-plugin.yml")

	var cmd = fmt.Sprintf("%s create -f %s", kubectl, pluginFile)
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type CheckGpuStatus struct {
	common.KubeAction
}

func (t *CheckGpuStatus) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	cmd := fmt.Sprintf("%s get pod  -n gpu-system -l 'app=orionx-container-runtime' -o jsonpath='{.items[*].status.phase}'", kubectl)

	rphase, _ := runtime.GetRunner().SudoCmdExt(cmd, false, false)
	if rphase == "Running" {
		return nil
	}
	return fmt.Errorf("GPU Container State is Pending")

}

type InstallGPUShared struct {
	common.KubeAction
}

func (t *InstallGPUShared) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)

	fileName := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir, "nvshare-system.yaml")
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectl, fileName), false, true); err != nil {
		return err
	}

	fileName = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir, "nvshare-system-quotas.yaml")
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectl, fileName), false, true); err != nil {
		return err
	}

	fileName = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir, "device-plugin.yaml")
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectl, fileName), false, true); err != nil {
		return err
	}

	fileName = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir, cc.GpuDir, "scheduler.yaml")
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectl, fileName), false, true); err != nil {
		return err
	}

	return nil
}
