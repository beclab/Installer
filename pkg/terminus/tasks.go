package terminus

import (
	"context"
	"fmt"
	"io/fs"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"time"

	bootstraptpl "bytetrade.io/web3os/installer/pkg/bootstrap/os/templates"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/terminus/templates"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	uninstalltemplate "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"

	"github.com/pkg/errors"
)

type GetTerminusVersion struct {
}

func (t *GetTerminusVersion) Execute() (string, error) {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return "", fmt.Errorf("kubectl not found, Olares might not be installed.")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf("%s get terminus -o jsonpath='{.items[*].spec.version}'", kubectlpath))
	cmd.WaitDelay = 3 * time.Second
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get Olares version failed")
	}

	if version := string(output); version == "" {
		return "", fmt.Errorf("Olares might not be installed.")
	} else {
		return version, nil
	}
}

type CheckKeyPodsRunning struct {
	common.KubeAction
}

func (t *CheckKeyPodsRunning) Execute(runtime connector.Runtime) error {
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load kubeconfig")
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create kube client")
	}
	pods, err := kubeClient.CoreV1().Pods(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list pods")
	}
	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Namespace, "user-space") ||
			strings.HasPrefix(pod.Namespace, "user-system") ||
			pod.Namespace == "os-system" {
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("pod %s/%s is not running", pod.Namespace, pod.Name)
			}
			if len(pod.Status.ContainerStatuses) == 0 {
				return fmt.Errorf("pod %s/%s has no container statuses yet", pod.Namespace, pod.Name)
			}
			for _, cStatus := range pod.Status.ContainerStatuses {
				if cStatus.State.Running == nil {
					return fmt.Errorf("container %s in pod %s/%s is not running", cStatus.Name, pod.Namespace, pod.Name)
				}
			}
		}
	}
	return nil
}

type CheckPodsRunning struct {
	common.KubeAction
	labels map[string][]string
}

func (c *CheckPodsRunning) Execute(runtime connector.Runtime) error {
	if c.labels == nil {
		return nil
	}

	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	for ns, labels := range c.labels {
		for _, label := range labels {
			var cmd = fmt.Sprintf("%s get pod -n %s -l '%s' -o jsonpath='{.items[*].status.phase}'", kubectl, ns, label)
			phase, err := runtime.GetRunner().Host.SudoCmdContext(ctx, cmd, false, false)
			if err != nil {
				return fmt.Errorf("pod status invalid, namespace: %s, label: %s, waiting ...", ns, label)
			}

			if phase != "Running" {
				logger.Infof("pod in namespace: %s, label: %s, current phase: %s, waiting ...", ns, label, phase)
				return fmt.Errorf("pod is %s, namespace: %s, label: %s, waiting ...", phase, ns, label)
			}
		}
	}

	return nil
}

type Download struct {
	common.KubeAction
	Version        string
	BaseDir        string
	DownloadCdnUrl string
}

func (t *Download) Execute(runtime connector.Runtime) error {
	if t.KubeConf.Arg.TerminusVersion == "" {
		return errors.New("unknown version to download")
	}

	var fetchMd5 = fmt.Sprintf("curl -sSfL %s/install-wizard-v%s.md5sum.txt |awk '{print $1}'", t.DownloadCdnUrl, t.Version)
	md5sum, err := runtime.GetRunner().Host.Cmd(fetchMd5, false, false)
	if err != nil {
		return errors.New("get md5sum failed")
	}

	var osArch = runtime.GetSystemInfo().GetOsArch()
	var osType = runtime.GetSystemInfo().GetOsType()
	var osVersion = runtime.GetSystemInfo().GetOsVersion()
	var osPlatformFamily = runtime.GetSystemInfo().GetOsPlatformFamily()
	var baseDir = runtime.GetBaseDir()
	var prePath = path.Join(baseDir, "versions")
	var wizard = files.NewKubeBinary("install-wizard", osArch, osType, osVersion, osPlatformFamily, t.Version, prePath, t.DownloadCdnUrl)
	wizard.CheckMd5Sum = true
	wizard.Md5sum = md5sum

	if err := wizard.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", wizard.FileName)
	}

	var exists = util.IsExist(wizard.Path())
	if exists {
		if err := wizard.Md5Check(); err == nil {
			// file exists, re-unpack
			return util.Untar(wizard.Path(), wizard.BaseDir)
		} else {
			logger.Error(err)
		}

		util.RemoveFile(wizard.Path())
	}

	logger.Infof("%s downloading %s %s ...", common.LocalHost, wizard.ID, wizard.Version)
	if err := wizard.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", wizard.ID, wizard.Url, err)
	}

	return util.Untar(wizard.Path(), wizard.BaseDir)
}

func copyWizard(wizardPath string, np string, runtime connector.Runtime) {
	if util.IsExist(np) {
		util.RemoveDir(np)
	} else {
		// util.Mkdir(np)
	}
	_, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("cp -a %s %s", wizardPath, np), false, false)
	if err != nil {
		logger.Errorf("copy -a %s to %s failed", wizardPath, np)
	}
}

type DownloadFullInstaller struct {
	common.KubeAction
}

func (t *DownloadFullInstaller) Execute(runtime connector.Runtime) error {

	return nil
}

type PrepareFinished struct {
	common.KubeAction
}

func (t *PrepareFinished) Execute(runtime connector.Runtime) error {
	var preparedFile = filepath.Join(runtime.GetBaseDir(), common.TerminusStateFilePrepared)
	return util.WriteFile(preparedFile, []byte(t.KubeConf.Arg.TerminusVersion), cc.FileMode0644)
	// if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("touch %s", preparedFile), false, true); err != nil {
	// 	return err
	// }
	// return nil
}

type CheckPrepared struct {
	common.KubeAction
	Force bool
}

func (t *CheckPrepared) Execute(runtime connector.Runtime) error {
	var preparedPath = filepath.Join(runtime.GetBaseDir(), common.TerminusStateFilePrepared)

	if utils.IsExist(preparedPath) {
		t.PipelineCache.Set(common.CachePreparedState, true)
	} else if t.Force {
		return errors.New("Olares dependencies is not prepared, refuse to continue")
	}

	return nil
}

type CheckInstalled struct {
	common.KubeAction
	Force bool
}

func (t *CheckInstalled) Execute(runtime connector.Runtime) error {
	var installedPath = filepath.Join(runtime.GetBaseDir(), common.TerminusStateFileInstalled)

	if utils.IsExist(installedPath) {
		t.PipelineCache.Set(common.CacheInstalledState, true)
	} else if t.Force {
		return errors.New("Olares is not installed, refuse to continue")
	}

	return nil
}

type GenerateOlaresUninstallScript struct {
	common.KubeAction
}

func (t *GenerateOlaresUninstallScript) Execute(runtime connector.Runtime) error {
	filePath := path.Join(runtime.GetBaseDir(), uninstalltemplate.OlaresUninstallScriptValues.Name())
	uninstallPath := path.Join("/usr/local/bin", uninstalltemplate.OlaresUninstallScriptValues.Name())
	data := util.Data{
		"BaseDir": runtime.GetBaseDir(),
		"Phase":   "install",
		"Version": t.KubeConf.Arg.TerminusVersion,
	}

	uninstallScriptStr, err := util.Render(uninstalltemplate.OlaresUninstallScriptValues, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render uninstall template failed")
	}

	if err := util.WriteFile(filePath, []byte(uninstallScriptStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write uninstall %s failed", filePath))
	}

	if err := runtime.GetRunner().Host.SudoScp(filePath, uninstallPath); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", filePath, uninstallPath))
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf %s", filePath), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove file %s failed", filePath))
	}

	return nil
}

type InstallFinished struct {
	common.KubeAction
}

func (t *InstallFinished) Execute(runtime connector.Runtime) error {
	var content = fmt.Sprintf("%s %s", t.KubeConf.Arg.TerminusVersion, t.KubeConf.Arg.Kubetype)
	var phaseState = path.Join(runtime.GetBaseDir(), common.TerminusStateFileInstalled)
	if err := util.WriteFile(phaseState, []byte(content), cc.FileMode0644); err != nil {
		return err
	}
	return nil
}

type DeleteWizardFiles struct {
	common.KubeAction
}

func (d *DeleteWizardFiles) Execute(runtime connector.Runtime) error {
	var locations = []string{
		path.Join(runtime.GetInstallerDir(), cc.BuildFilesCacheDir),
		runtime.GetWorkDir(),
		path.Join(runtime.GetInstallerDir(), cc.LogsDir, cc.InstallLogFile),
	}

	for _, location := range locations {
		if util.IsExist(location) {
			runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf %s", location), false, true)
		}
	}
	return nil
}

type SystemctlCommand struct {
	common.KubeAction
	UnitNames           []string
	Command             string
	DaemonReloadPreExec bool
}

func (a *SystemctlCommand) Execute(runtime connector.Runtime) error {
	if a.DaemonReloadPreExec {
		if _, err := runtime.GetRunner().Host.SudoCmd("systemctl daemon-reload", false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "systemctl reload failed")
		}
	}
	for _, unitName := range a.UnitNames {
		cmd := fmt.Sprintf("systemctl %s %s", a.Command, unitName)
		if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
			return errors.Wrapf(err, "failed to execute command: %s", cmd)
		}
	}

	return nil
}

// PrepareFilesForETCDIPChange simply copies the current CA cert and key
// to the work directory of Installer,
// so that after the regeneration is done,
// all other certs and keys is replaced with a new one
// but issued by the same CA
type PrepareFilesForETCDIPChange struct {
	common.KubeAction
}

func (a *PrepareFilesForETCDIPChange) Execute(runtime connector.Runtime) error {
	srcCertsDir := "/etc/ssl/etcd/ssl"
	dstCertsDir := filepath.Join(runtime.GetWorkDir(), "/pki/etcd")
	if err := util.RemoveDir(dstCertsDir); err != nil {
		return errors.Wrap(err, "failed to clear work directory for etcd certs")
	}
	if err := util.Mkdir(dstCertsDir); err != nil {
		return errors.Wrap(err, "failed to create work directory for etcd certs")
	}
	return filepath.WalkDir(srcCertsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(filepath.Base(path), "ca") {
			if err := util.CopyFileToDir(path, dstCertsDir); err != nil {
				return err
			}
		}
		return nil
	})
}

type PrepareFilesForK8sIPChange struct {
	common.KubeAction
}

func (a *PrepareFilesForK8sIPChange) Execute(runtime connector.Runtime) error {
	k8sConfDir := "/etc/kubernetes"
	return filepath.WalkDir(k8sConfDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".conf" {
			logger.Debugf("removing %s", path)
			return util.RemoveFile(path)
		}
		if filepath.Base(path) == "kubeadm-config.yaml" {
			logger.Debugf("removing %s", path)
			return util.RemoveFile(path)
		}
		if strings.Contains(path, "pki") && !strings.HasPrefix(filepath.Base(path), "ca") {
			logger.Debugf("removing %s", path)
			return util.RemoveFile(path)
		}
		return nil
	})
}

type RegenerateFilesForK8sIPChange struct {
	common.KubeAction
}

func (a *RegenerateFilesForK8sIPChange) Execute(runtime connector.Runtime) error {
	initCmd := "/usr/local/bin/kubeadm init --config=/etc/kubernetes/kubeadm-config.yaml --skip-phases=preflight,mark-control-plane,bootstrap-token,addon"

	if _, err := runtime.GetRunner().Host.SudoCmd(initCmd, false, false); err != nil {
		return err
	}
	return nil
}

type DeleteAllPods struct {
	common.KubeAction
}

func (a *DeleteAllPods) Execute(runtime connector.Runtime) error {
	kubectlpath, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}
	var cmd = fmt.Sprintf("%s delete pod --all --all-namespaces", kubectlpath)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type DeletePodsUsingHostIP struct {
	common.KubeAction
}

func (a *DeletePodsUsingHostIP) Execute(runtime connector.Runtime) error {
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load kubeconfig")
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create kube client")
	}

	targetPods, err := getPodsUsingHostIP(kubeClient)
	if err != nil {
		return errors.Wrap(err, "failed to get pods using host IP")
	}
	a.PipelineCache.Set(common.CacheCountPodsUsingHostIP, len(targetPods))
	for _, pod := range targetPods {
		logger.Infof("restarting pod %s/%s that's using host IP", pod.Namespace, pod.Name)
		err = kubeClient.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
		if err != nil && !kerrors.IsNotFound(err) {
			return errors.Wrap(err, "failed to delete pod")
		}
	}

	// try our best to wait for the pods to be actually deleted
	// to avoid the next module getting the pods with a still running phase
	err = waitForPodsToBeGone(kubeClient, targetPods, 3*time.Minute)
	if err != nil {
		logger.Warnf("failed to wait for pods to be gone: %v, will delay and skip", err)
		time.Sleep(60 * time.Second)
		return nil
	}

	return nil
}

type WaitForPodsUsingHostIPRecreate struct {
	common.KubeAction
}

func (a *WaitForPodsUsingHostIPRecreate) Execute(runtime connector.Runtime) error {
	count, ok := a.PipelineCache.GetMustInt(common.CacheCountPodsUsingHostIP)
	if !ok {
		return errors.New("failed to get the count of pods using host IP")
	}
	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load kubeconfig")
	}
	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to create kube client")
	}

	targetPods, err := getPodsUsingHostIP(kubeClient)
	if err != nil {
		return errors.Wrap(err, "failed to get pods using host IP")
	}
	if len(targetPods) < count {
		return errors.New("waiting for pods using host IP to be recreated")
	}
	return nil
}

func getPodsUsingHostIP(kubeClient kubernetes.Interface) ([]*corev1.Pod, error) {
	pods, err := kubeClient.CoreV1().Pods(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil || len(pods.Items) == 0 {
		return nil, errors.Wrap(err, "failed to list pods")
	}
	var targetPods []*corev1.Pod
	for _, pod := range pods.Items {
		if podIsUsingHostIP(&pod) {
			targetPods = append(targetPods, &pod)
		}
	}
	return targetPods, nil
}

func podIsUsingHostIP(pod *corev1.Pod) bool {
	if pod == nil {
		return false
	}
	if pod.Spec.HostNetwork == true {
		return true
	}
	var allContainers []corev1.Container
	allContainers = append(allContainers, pod.Spec.Containers...)
	allContainers = append(allContainers, pod.Spec.InitContainers...)
	for _, container := range allContainers {
		for _, env := range container.Env {
			if env.ValueFrom != nil && env.ValueFrom.FieldRef != nil && env.ValueFrom.FieldRef.FieldPath == "status.hostIP" {
				return true
			}
		}
	}
	return false
}

func waitForPodsToBeGone(kubeClient *kubernetes.Clientset, pods []*corev1.Pod, timeout time.Duration) error {
	for _, pod := range pods {
		pod, err := kubeClient.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		if kerrors.IsNotFound(err) {
			continue
		}
		if err != nil {
			return errors.Wrap(err, "failed to check if pod exists")
		}
		watchOptions := metav1.ListOptions{
			FieldSelector:   fields.OneTermEqualSelector("metadata.name", pod.Name).String(),
			ResourceVersion: pod.GetResourceVersion(),
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		watcher, err := kubeClient.CoreV1().Pods(pod.Namespace).Watch(ctx, watchOptions)
		if err != nil {
			return errors.Wrapf(err, "failed to watch pod %s for deletion", client.ObjectKeyFromObject(pod).String())
		}
		defer watcher.Stop()
		watchChan := watcher.ResultChan()
		for {
			event, ok := <-watchChan
			if !ok || event.Type == watch.Error {
				if !ok {
					err = errors.New("watch channel closed unexpectedly")
				} else {
					err = kerrors.FromObject(event.Object)
				}
				return errors.Wrapf(err, "failed to watch pod %s for deletion", client.ObjectKeyFromObject(pod).String())
			}
			if event.Type == watch.Deleted {
				break
			}
		}
	}
	return nil
}

type UpdateKubeKeyHosts struct {
	common.KubeAction
}

func (a *UpdateKubeKeyHosts) Execute(runtime connector.Runtime) error {

	scriptPath := filepath.Join(runtime.GetWorkDir(), "change-ip-scripts", "update-kubekey-hosts.sh")
	tplAction := &action.Template{
		Name:     "GenerateHostsUpdateScript",
		Template: templates.UpdateKKHostsScriptTmpl,
		Dst:      scriptPath,
		Data: util.Data{
			"Hosts": bootstraptpl.GenerateHosts(runtime, a.KubeConf),
		},
	}
	if err := tplAction.Execute(runtime); err != nil {
		return errors.Wrapf(err, fmt.Sprintf("failed to generate update hosts script: %s", scriptPath))
	}
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("chmod +x %s", scriptPath), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to chmod +x update hosts script")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(scriptPath, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "failed to run update hosts script")
	}
	return nil
}

type CheckTerminusStateInHost struct {
	common.KubeAction
}

func (a *CheckTerminusStateInHost) Execute(runtime connector.Runtime) error {
	si := runtime.GetSystemInfo()
	var kubectlCMD string
	var kubectlCMDDefaultArgs []string
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if si.IsDarwin() {
		minikube, err := util.GetCommand(common.CommandMinikube)
		if err != nil {
			return errors.Wrap(err, "failed to get minikube command")
		}
		_, _, err = util.Exec(ctx, fmt.Sprintf("%s update-context -p %s", minikube, a.KubeConf.Arg.MinikubeProfile), false, true)
		if err != nil {
			fmt.Printf("failed to update minikube context to %s, is it up and running?", a.KubeConf.Arg.MinikubeProfile)
			os.Exit(1)
		}
		kubectlCMD, err = util.GetCommand(common.CommandKubectl)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "kubectl not found")
		}
	} else if si.IsWindows() {
		kubectlCMD = "cmd"
		kubectlCMDDefaultArgs = []string{"/C", "wsl", "-d", a.KubeConf.Arg.WSLDistribution, "-u", "root", common.CommandKubectl}
	}

	getTerminusArgs := []string{"get", "terminus"}

	getTerminusCMD := exec.CommandContext(ctx, kubectlCMD, append(kubectlCMDDefaultArgs, getTerminusArgs...)...)
	getTerminusCMD.WaitDelay = 3 * time.Second
	output, err := getTerminusCMD.CombinedOutput()
	if err != nil {
		logger.Debugf("failed to run command %v, ouput: %s, err: %s", getTerminusCMD, output, err)
		fmt.Println("failed to check the existence of terminus, is it installed and running?")
		os.Exit(1)
	}

	return nil
}

type GetPublicNetworkInfo struct {
	common.KubeAction
}

func (p *GetPublicNetworkInfo) Execute(runtime connector.Runtime) error {
	osPublicIPs, err := util.GetPublicIPsFromOS()
	if err != nil {
		return errors.Wrap(err, "failed to get public IPs from OS")
	}
	if len(osPublicIPs) > 0 {
		logger.Info("detected public IP addresses on local network interface")
		p.KubeConf.Arg.PublicNetworkInfo.OSPublicIPs = osPublicIPs
		return nil
	}

	if util.IsOnAWSEC2() {
		logger.Info("on AWS EC2 instance, will try to check if a public IP address is bound")
		awsPublicIP, err := util.GetPublicIPFromAWSIMDS()
		if err != nil {
			return errors.Wrap(err, "failed to get public IP from AWS")
		}
		if awsPublicIP != nil {
			logger.Info("retrieved public IP addresses from IMDS")
			p.KubeConf.Arg.PublicNetworkInfo.AWSPublicIP = awsPublicIP
			return nil
		}
	}

	return nil
}
