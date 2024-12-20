package gpu

import (
	"context"
	"fmt"
	"path"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	k3sGpuTemplates "bytetrade.io/web3os/installer/pkg/gpu/templates"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"github.com/pkg/errors"
)

type CheckWslGPU struct {
}

func (t *CheckWslGPU) CheckNvidiaSmiFileExists() bool {
	var nvidiaSmiFile = "/usr/lib/wsl/lib/nvidia-smi"
	if !util.IsExist(nvidiaSmiFile) {
		return false
	}
	return true
}

func (t *CheckWslGPU) Execute(runtime *common.KubeRuntime) {
	if !runtime.GetSystemInfo().IsWsl() {
		return
	}
	exists := t.CheckNvidiaSmiFileExists()
	if !exists {
		return
	}

	stdout, _, err := util.Exec(context.Background(), "/usr/lib/wsl/lib/nvidia-smi -L|grep 'NVIDIA'|grep UUID", false, false)
	if err != nil {
		logger.Errorf("nvidia-smi not found")
		return
	}
	if stdout == "" {
		return
	}

	runtime.Arg.SetGPU(true, true)
}

type InstallCudaDeps struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallCudaDeps) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var fileId = fmt.Sprintf("%s-%s_cuda-keyring_%s-1",
		strings.ToLower(systemInfo.GetOsPlatformFamily()), systemInfo.GetOsVersion(),
		kubekeyapiv1alpha2.DefaultCudaKeyringVersion)

	cudakeyring, err := t.Manifest.Get(fileId)
	if err != nil {
		return err
	}

	path := cudakeyring.FilePath(t.BaseDir)
	var exists = util.IsExist(path)
	if !exists {
		return fmt.Errorf("Failed to find %s binary in %s", cudakeyring.Filename, path)
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("dpkg -i %s", path), false, true); err != nil {
		return err
	}

	return nil
}

type InstallCudaDriver struct {
	common.KubeAction
}

func (t *InstallCudaDriver) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Host.SudoCmd("apt-get update", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get update")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("apt-get -y install nvidia-kernel-open-550", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-kernel-open-550")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("apt-get -y install nvidia-driver-550", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-driver-550")
	}

	// if _, err := runtime.GetRunner().Host.SudoCmd("apt-get -y install cuda-12-1", false, true); err != nil {
	// 	return errors.Wrap(errors.WithStack(err), "Failed to apt-get install cuda-12-1")
	// }

	return nil
}

type UpdateCudaSource struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *UpdateCudaSource) Execute(runtime connector.Runtime) error {
	// only for ubuntu20.04  ubunt22.04
	systemInfo := runtime.GetSystemInfo()

	var version string
	if strings.Contains(systemInfo.GetOsVersion(), "22.") {
		version = "22.04"
	} else {
		version = "20.04"
	}

	var cmd string
	gpgkey, err := t.Manifest.Get("gpgkey")
	if err != nil {
		return err
	}

	keyPath := gpgkey.FilePath(t.BaseDir)

	if !util.IsExist(keyPath) {
		return fmt.Errorf("Failed to find %s binary in %s", gpgkey.Filename, keyPath)
	}

	cmd = fmt.Sprintf("apt-key add %s", keyPath)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}

	if strings.Contains(systemInfo.GetOsVersion(), "24.") {
		return nil
	}

	var fileId = fmt.Sprintf("%s_%s_libnvidia-container.list",
		strings.ToLower(systemInfo.GetOsPlatformFamily()), version)

	libnvidia, err := t.Manifest.Get(fileId)
	if err != nil {
		return err
	}

	libPath := libnvidia.FilePath(t.BaseDir)

	if !util.IsExist(libPath) {
		return fmt.Errorf("Failed to find %s binary in %s", libnvidia.Filename, libPath)
	}

	cmd = fmt.Sprintf("cp %s %s", libPath, "/etc/apt/sources.list.d/")
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type InstallNvidiaContainerToolkit struct {
	common.KubeAction
}

func (t *InstallNvidiaContainerToolkit) Execute(runtime connector.Runtime) error {
	logger.Debugf("install nvidia-container-toolkit")
	if _, err := runtime.GetRunner().Host.SudoCmd("apt-get update && sudo apt-get install -y nvidia-container-toolkit jq", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-container-toolkit")
	}
	return nil
}

type PatchK3sDriver struct { // patch k3s on wsl
	common.KubeAction
}

func (t *PatchK3sDriver) Execute(runtime connector.Runtime) error {
	var cmd = "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1"
	driverPath, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true)
	if err != nil {
		return err
	}

	if driverPath == "" {
		logger.Infof("cuda driver not found")
		return nil
	} else {
		logger.Infof("cuda driver found: %s", driverPath)
	}

	templateStr, err := util.Render(k3sGpuTemplates.K3sCudaFixValues, nil)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("render template %s failed", k3sGpuTemplates.K3sCudaFixValues.Name()))
	}

	var fixName = "cuda_lib_fix.sh"
	var fixPath = path.Join(runtime.GetBaseDir(), cc.PackageCacheDir, "gpu", "cuda_lib_fix.sh")
	if err := util.WriteFile(fixPath, []byte(templateStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write file %s failed", fixPath))
	}

	var dstName = path.Join(common.BinDir, fixName)
	if err := runtime.GetRunner().Host.SudoScp(fixPath, dstName); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fixPath, dstName))
	}

	cmd = fmt.Sprintf("echo 'ExecStartPre=-/usr/local/bin/%s' >> /etc/systemd/system/k3s.service", fixName)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	return nil
}

type RestartContainerd struct {
	common.KubeAction
}

func (t *RestartContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Host.SudoCmd("nvidia-ctk runtime configure --runtime=containerd --set-as-default --config-source=command", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to nvidia-ctk runtime configure")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl restart containerd", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart containerd")
	}
	return nil
}

type InstallPlugin struct {
	common.KubeAction
}

func (t *InstallPlugin) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var pluginFile = path.Join(runtime.GetInstallerDir(), "deploy", "nvidia-device-plugin.yml")
	if !util.IsExist(pluginFile) {
		logger.Errorf("plugin file not exist: %s", pluginFile)
		return nil
	}
	var cmd = fmt.Sprintf("%s create -f %s", kubectlpath, pluginFile)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type CheckGpuStatus struct {
	common.KubeAction
}

func (t *CheckGpuStatus) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	cmd := fmt.Sprintf("%s get pod  -n kube-system -l 'name=nvidia-device-plugin-ds' -o jsonpath='{.items[*].status.phase}'", kubectlpath)

	rphase, _ := runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	if rphase == "Running" {
		return nil
	}
	return fmt.Errorf("GPU Container State is Pending")
}

type InstallGPUShared struct {
	common.KubeAction
}

func (t *InstallGPUShared) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	var pluginPath = runtime.GetInstallerDir()
	var fileName = path.Join(pluginPath, "deploy", "nvshare-system.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply nvshare-system.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "nvshare-system-quotas.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply nvshare-system-quotas.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "device-plugin.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply device-plugin.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "scheduler.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply scheduler.yaml")
	}

	return nil
}

type GetCudaVersion struct {
	common.KubeAction
}

func (g *GetCudaVersion) Execute(runtime connector.Runtime) error {
	var nvidiaSmiFile string
	var systemInfo = runtime.GetSystemInfo()

	switch {
	case systemInfo.IsWsl():
		nvidiaSmiFile = "/usr/lib/wsl/lib/nvidia-smi"
	default:
		nvidiaSmiFile = "/usr/bin/nvidia-smi"
	}

	if !util.IsExist(nvidiaSmiFile) {
		logger.Info("nvidia-smi not exists")
		return nil
	}

	var cudaVersion string
	res, err := runtime.GetRunner().Host.Cmd(fmt.Sprintf("%s --version", nvidiaSmiFile), false, true)
	if err != nil {
		logger.Errorf("get cuda version error %v", err)
		return nil
	}

	lines := strings.Split(res, "\n")

	if lines == nil || len(lines) == 0 {
		return nil
	}
	for _, line := range lines {
		if strings.Contains(line, "CUDA Version") {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				break
			}
			cudaVersion = strings.TrimSpace(parts[1])
		}
	}
	if cudaVersion != "" {
		common.TerminusGlobalEnvs["CUDA_VERSION"] = cudaVersion
	}

	return nil
}
