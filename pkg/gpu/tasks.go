package gpu

import (
	"bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/bootstrap/precheck"
	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	k3sGpuTemplates "bytetrade.io/web3os/installer/pkg/gpu/templates"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
	criconfig "github.com/containerd/containerd/pkg/cri/config"
	cdsrvconfig "github.com/containerd/containerd/services/server/config"
	"github.com/pelletier/go-toml"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
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
	var cudaKeyringVersion string
	var osVersion string
	switch {
	case systemInfo.IsUbuntu():
		cudaKeyringVersion = v1alpha2.CudaKeyringVersion1_0
		if systemInfo.IsUbuntuVersionEqual(connector.Ubuntu24) {
			cudaKeyringVersion = v1alpha2.CudaKeyringVersion1_1
			osVersion = "24.04"
		} else if systemInfo.IsUbuntuVersionEqual(connector.Ubuntu22) {
			osVersion = "22.04"
		} else {
			osVersion = "20.04"
		}
	case systemInfo.IsDebian():
		cudaKeyringVersion = v1alpha2.CudaKeyringVersion1_1
		if systemInfo.IsDebianVersionEqual(connector.Debian12) {
			osVersion = connector.Debian12.String()
		} else {
			osVersion = connector.Debian11.String()
		}
	}
	var fileId = fmt.Sprintf("%s-%s_cuda-keyring_%s-1",
		strings.ToLower(systemInfo.GetOsPlatformFamily()), osVersion, cudaKeyringVersion)

	cudakeyring, err := t.Manifest.Get(fileId)
	if err != nil {
		return err
	}

	path := cudakeyring.FilePath(t.BaseDir)
	var exists = util.IsExist(path)
	if !exists {
		return fmt.Errorf("Failed to find %s binary in %s", cudakeyring.Filename, path)
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("dpkg -i --force all %s", path), false, true); err != nil {
		return err
	}

	return nil
}

type InstallCudaDriver struct {
	common.KubeAction
}

func (t *InstallCudaDriver) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("apt-get update", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get update")
	}

	if runtime.GetSystemInfo().IsDebian() {
		_, err := runtime.GetRunner().SudoCmd("apt-get -y install nvidia-open", false, true)
		return errors.Wrap(err, "failed to apt-get install nvidia-open")
	}

	if _, err := runtime.GetRunner().SudoCmd("apt-get -y install nvidia-kernel-open-550", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-kernel-open-550")
	}

	if _, err := runtime.GetRunner().SudoCmd("apt-get -y install nvidia-driver-550", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-driver-550")
	}

	// if _, err := runtime.GetRunner().SudoCmd("apt-get -y install cuda-12-1", false, true); err != nil {
	// 	return errors.Wrap(errors.WithStack(err), "Failed to apt-get install cuda-12-1")
	// }

	return nil
}

type UpdateCudaSource struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *UpdateCudaSource) Execute(runtime connector.Runtime) error {
	var cmd string
	gpgkey, err := t.Manifest.Get("libnvidia-gpgkey")
	if err != nil {
		return err
	}

	keyPath := gpgkey.FilePath(t.BaseDir)

	if !util.IsExist(keyPath) {
		return fmt.Errorf("Failed to find %s binary in %s", gpgkey.Filename, keyPath)
	}

	cmd = fmt.Sprintf("apt-key add %s", keyPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}

	libnvidia, err := t.Manifest.Get("libnvidia-container.list")
	if err != nil {
		return err
	}

	libPath := libnvidia.FilePath(t.BaseDir)

	if !util.IsExist(libPath) {
		return fmt.Errorf("Failed to find %s binary in %s", libnvidia.Filename, libPath)
	}

	// remove any conflicting libnvidia-container.list
	_, err = runtime.GetRunner().SudoCmd("rm -rf /etc/apt/sources.list.d/*nvidia-container*.list", false, false)
	if err != nil {
		return err
	}

	dstPath := filepath.Join("/etc/apt/sources.list.d", filepath.Base(libPath))
	cmd = fmt.Sprintf("cp %s %s", libPath, dstPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
		return err
	}

	mirrorRepo := os.Getenv(common.ENV_NVIDIA_CONTAINER_REPO_MIRROR)
	if mirrorRepo == "" {
		return nil
	}
	mirrorRepoRawURL := mirrorRepo
	if !strings.HasPrefix(mirrorRepoRawURL, "http") {
		mirrorRepoRawURL = "https://" + mirrorRepoRawURL
	}
	mirrorRepoURL, err := url.Parse(mirrorRepoRawURL)
	if err != nil || mirrorRepoURL.Host == "" {
		return fmt.Errorf("invalid mirror for nvidia container: %s", mirrorRepo)
	}
	cmd = fmt.Sprintf("sed -i 's#nvidia.github.io#%s#g' %s", mirrorRepoURL.Host, dstPath)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to switch nvidia container repo to mirror site")
	}
	return nil
}

type InstallNvidiaContainerToolkit struct {
	common.KubeAction
}

func (t *InstallNvidiaContainerToolkit) Execute(runtime connector.Runtime) error {
	logger.Debugf("install nvidia-container-toolkit")
	if _, err := runtime.GetRunner().SudoCmd("apt-get update && sudo apt-get install -y nvidia-container-toolkit jq", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get install nvidia-container-toolkit")
	}
	return nil
}

type PatchK3sDriver struct { // patch k3s on wsl
	common.KubeAction
}

func (t *PatchK3sDriver) Execute(runtime connector.Runtime) error {
	if !runtime.GetSystemInfo().IsWsl() {
		return nil
	}
	var cmd = "find /usr/lib/wsl/drivers/ -name libcuda.so.1.1|head -1"
	driverPath, err := runtime.GetRunner().SudoCmd(cmd, false, true)
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
	if err := runtime.GetRunner().SudoScp(fixPath, dstName); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", fixPath, dstName))
	}

	cmd = fmt.Sprintf("echo 'ExecStartPre=-/usr/local/bin/%s' >> /etc/systemd/system/k3s.service", fixName)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().SudoCmd(dstName, false, false); err != nil {
		return errors.Wrap(err, "failed to apply CUDA patch for WSL")
	}

	return nil
}

type ConfigureContainerdRuntime struct {
	common.KubeAction
}

func (t *ConfigureContainerdRuntime) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("nvidia-ctk runtime configure --runtime=containerd --set-as-default --config-source=command", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to nvidia-ctk runtime configure")
	}

	return nil
}

type RestartContainerd struct {
	common.KubeAction
}

func (t *RestartContainerd) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart containerd", false, true); err != nil {
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
	var cmd = fmt.Sprintf("%s apply -f %s", kubectlpath, pluginFile)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, true); err != nil {
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

	rphase, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
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
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply nvshare-system.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "nvshare-system-quotas.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply nvshare-system-quotas.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "device-plugin.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apply device-plugin.yaml")
	}

	fileName = path.Join(pluginPath, "deploy", "scheduler.yaml")
	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s apply -f %s", kubectlpath, fileName), false, true); err != nil {
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
	res, err := runtime.GetRunner().Cmd(fmt.Sprintf("%s --version", nvidiaSmiFile), false, true)
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

type UpdateNodeLabels struct {
	common.KubeAction
	precheck.CudaCheckTask
}

func (u *UpdateNodeLabels) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	gpuInfo, installed, err := utils.ExecNvidiaSmi(runtime)
	if err != nil {
		return err
	}

	if !installed {
		logger.Info("nvidia-smi not exists")
		return nil
	}

	supported := "false"

	err = u.CudaCheckTask.Execute(runtime)
	switch {
	case err == precheck.ErrCudaInstalled:
		supported = "true"
	case err == precheck.ErrUnsupportedCudaVersion:
		// bypass
	case err != nil:
		return err
	case err == nil:
		// impossible
		logger.Warn("check impossible")
	}

	return UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), &gpuInfo.DriverVersion, &gpuInfo.CudaVersion, &supported)
}

type RemoveNodeLabels struct {
	common.KubeAction
}

func (u *RemoveNodeLabels) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	return UpdateNodeGpuLabel(context.Background(), client.Kubernetes(), nil, nil, nil)
}

// update k8s node labels gpu.bytetrade.io/driver and gpu.bytetrade.io/cuda.
// if labels are not exists, create it.
func UpdateNodeGpuLabel(ctx context.Context, client kubernetes.Interface, driver, cuda *string, supported *string) error {
	// get node name from hostname
	nodeName, err := os.Hostname()
	if err != nil {
		logger.Error("get hostname error, ", err)
		return err
	}

	node, err := client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Error("get node error, ", err)
		return err
	}

	labels := node.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	update := false
	for _, label := range []struct {
		key   string
		value *string
	}{
		{GpuDriverLabel, driver},
		{GpuCudaLabel, cuda},
		{GpuCudaSupportedLabel, supported},
	} {
		old, ok := labels[label.key]
		switch {
		case ok && label.value == nil: // delete label
			delete(labels, label.key)
			update = true

		case ok && *label.value != "" && old != *label.value: // update label
			labels[label.key] = *label.value
			update = true

		case !ok && label.value != nil && *label.value != "": // create label
			labels[label.key] = *label.value
			update = true
		}
	}

	if update {
		node.SetLabels(labels)
		safeString := func(s *string) string {
			if s == nil {
				return "nil"
			}
			return *s
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			logger.Infof("updating node gpu labels, %s, %s", safeString(driver), safeString(cuda))
			_, err := client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
			return err
		})

		if err != nil {
			logger.Error("update node error, ", err)
			return err
		}
	}

	return nil
}

type RemoveContainerRuntimeConfig struct {
	common.KubeAction
}

func (t *RemoveContainerRuntimeConfig) Execute(runtime connector.Runtime) error {
	var configFile = "/etc/containerd/config.toml"
	var nvidiaRuntime = "nvidia"
	var criPluginUri = "io.containerd.grpc.v1.cri"

	if !util.IsExist(configFile) {
		logger.Infof("containerd config file not found")
		return nil
	}

	config := &cdsrvconfig.Config{}
	err := cdsrvconfig.LoadConfig(configFile, config)
	if err != nil {
		return fmt.Errorf("failed to load containerd config: %w", err)
	}
	plugins := config.Plugins[criPluginUri]
	var criConfig criconfig.PluginConfig
	if err := plugins.Unmarshal(&criConfig); err != nil {
		logger.Error("unmarshal cri config error: ", err)
		return err
	}

	// found nvidia runtime, remove it
	if _, ok := criConfig.ContainerdConfig.Runtimes[nvidiaRuntime]; ok {
		delete(criConfig.ContainerdConfig.Runtimes, nvidiaRuntime)
		criConfig.DefaultRuntimeName = "runc"

		// save config
		criConfigData, err := toml.Marshal(criConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal containerd cri plugin config: %w", err)
		}

		criPluginConfigTree, err := toml.LoadBytes(criConfigData)
		if err != nil {
			return fmt.Errorf("failed to load containerd cri plugin config: %w", err)
		}

		config.Plugins[criPluginUri] = *criPluginConfigTree

		// save config to file
		tmpConfigFile, err := os.OpenFile(configFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open minikube containerd config temp file for writing: %w", err)
		}
		defer tmpConfigFile.Close()
		if err := toml.NewEncoder(tmpConfigFile).Encode(config); err != nil {
			return fmt.Errorf("failed to write minikube containerd config temp file: %w", err)
		}

	}

	return nil
}

type UninstallNvidiaDrivers struct {
	common.KubeAction
}

func (t *UninstallNvidiaDrivers) Execute(runtime connector.Runtime) error {

	if _, err := runtime.GetRunner().SudoCmd("apt-get -y remove nvidia*", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get remove nvidia*")
	}

	if _, err := runtime.GetRunner().SudoCmd("apt-get -y remove libnvidia*", false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to apt-get remove libnvidia*")
	}

	logger.Infof("uninstall nvidia drivers success, please reboot the system to take effect if you reinstall the new nvidia drivers")
	return nil
}

type PrintGpuStatus struct {
	common.KubeAction
}

func (t *PrintGpuStatus) Execute(runtime connector.Runtime) error {
	gpuInfo, installed, err := utils.ExecNvidiaSmi(runtime)
	if err != nil {
		return err
	}

	if !installed {
		logger.Info("cuda not exists")
		return nil
	}

	logger.Infof("GPU Driver Version: %s", gpuInfo.DriverVersion)
	logger.Infof("CUDA Version: %s", gpuInfo.CudaVersion)

	return nil
}

type PrintPluginsStatus struct {
	common.KubeAction
}

func (t *PrintPluginsStatus) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	plugins, err := client.Kubernetes().CoreV1().Pods("kube-system").List(context.Background(), metav1.ListOptions{LabelSelector: "name=nvidia-device-plugin-ds"})
	if err != nil {
		logger.Error("get plugin status error, ", err)
		return err
	}

	if len(plugins.Items) == 0 {
		logger.Info("nvidia-device-plugin not exists")

	} else {
		for _, plugin := range plugins.Items {
			logger.Infof("nvidia-device-plugin status: %s", plugin.Status.Phase)
			break
		}
	}

	nvsharePlugins, err := client.Kubernetes().CoreV1().Pods("nvshare-system").List(context.Background(), metav1.ListOptions{LabelSelector: "name=nvshare-device-plugin"})
	if err != nil {
		logger.Error("get nvshare plugin status error, ", err)
		return err
	}

	if len(nvsharePlugins.Items) == 0 {
		logger.Info("nvshare-device-plugin not exists")

	} else {
		for _, plugin := range nvsharePlugins.Items {
			logger.Infof("nvshare-device-plugin status: %s", plugin.Status.Phase)
			break
		}
	}

	nvshareScheduler, err := client.Kubernetes().CoreV1().Pods("nvshare-system").List(context.Background(), metav1.ListOptions{LabelSelector: "name=nvshare-scheduler"})
	if err != nil {
		logger.Error("get nvshare scheduler status error, ", err)
	}

	if len(nvshareScheduler.Items) == 0 {
		logger.Info("nvshare-scheduler not exists")
	} else {
		for _, scheduler := range nvshareScheduler.Items {
			logger.Infof("nvshare-scheduler status: %s", scheduler.Status.Phase)
			break
		}
	}

	return nil
}

type RestartPlugin struct {
	common.KubeAction
}

func (t *RestartPlugin) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart ds nvshare-device-plugin -n nvshare-system", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart nvshare-device-plugin")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart ds nvshare-scheduler -n nvshare-system", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart nvshare-scheduler")
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s rollout restart ds nvidia-device-plugin-daemonset -n kube-system", kubectlpath), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "Failed to restart nvidia-device-plugin")
	}

	return nil
}
