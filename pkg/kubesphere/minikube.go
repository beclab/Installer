package kubesphere

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere/templates"
	mk "bytetrade.io/web3os/installer/pkg/version/minikube"
	"github.com/pkg/errors"
)

type CreateTerminus struct {
	common.KubeAction
}

func (t *CreateTerminus) Execute(runtime connector.Runtime) error {
	minikube, err := util.GetCommand(common.CommandMinikube)
	if err != nil {
		return fmt.Errorf("Please install minikube on your machine")
	}

	var cmd = fmt.Sprintf("%s profile list|grep '%s'|grep Running", minikube, t.KubeConf.Arg.MinikubeProfile)
	stdout, err := runtime.GetRunner().SudoCmdExt(cmd, false, true)
	if stdout != "" {
		return fmt.Errorf("minikube profile already exists")
	}

	cmd = fmt.Sprintf("%s start -p '%s' --kubernetes-version=v1.22.10 --network-plugin=cni --cni=calico --cpus='4' --memory='8g' --ports=30180:30180,443:443,80:80", minikube, t.KubeConf.Arg.MinikubeProfile)
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		return err
	}

	return nil
}

type CreateMinikubeModule struct {
	common.KubeModule
}

func (m *CreateMinikubeModule) Init() {
	m.Name = "CreateMinikube"

	createTerminus := &task.LocalTask{
		Name:   "Create",
		Action: new(CreateTerminus),
	}

	m.Tasks = []task.Interface{
		createTerminus,
	}
}

type UninstallMinikube struct {
	common.KubeAction
}

func (t *UninstallMinikube) Execute(runtime connector.Runtime) error {
	var minikubepath string
	var err error
	if minikubepath, err = util.GetCommand(common.CommandMinikube); err != nil || minikubepath == "" {
		return fmt.Errorf("minikube not found")
	}

	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("%s stop --all && %s delete --all", minikubepath, minikubepath), false, true); err != nil {
		return err
	}

	var phaseStats = []string{".installed", ".prepared"}
	for _, ps := range phaseStats {
		if util.IsExist(path.Join(runtime.GetBaseDir(), ps)) {
			util.RemoveFile(path.Join(runtime.GetBaseDir(), ps))
		}
	}
	return nil
}

type DeleteMinikubeModule struct {
	common.KubeModule
}

func (m *DeleteMinikubeModule) Init() {
	m.Name = "Uninstall"

	uninstallMinikube := &task.LocalTask{
		Name:   "Uninstall",
		Action: new(UninstallMinikube),
	}

	m.Tasks = []task.Interface{
		uninstallMinikube,
	}
}

type Download struct {
	common.KubeAction
}

func (t *Download) Execute(runtime connector.Runtime) error {
	var arch = runtime.GetRunner().Host.GetArch()

	var systemInfo = runtime.GetSystemInfo()
	var osType = systemInfo.GetOsType()
	var osVersion = systemInfo.GetOsVersion()
	var osPlatformFamily = systemInfo.GetOsPlatformFamily()
	helm := files.NewKubeBinary("helm", arch, osType, osVersion, osPlatformFamily, kubekeyapiv1alpha2.DefaultHelmVersion, runtime.GetWorkDir())

	if err := helm.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", helm.FileName)
	}

	var exists = util.IsExist(helm.Path())
	if exists {
		p := helm.Path()
		if err := helm.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		}
	}

	if !exists || helm.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, helm.ID, helm.Version)
		if err := helm.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", helm.ID, helm.Url, err)
		}
	}

	return nil
}

type DownloadMinikubeBinaries struct {
	common.KubeModule
}

func (m *DownloadMinikubeBinaries) Init() {
	m.Name = "DownloadMinikubeBinaries"

	downloadBinaries := &task.RemoteTask{
		Name:     "DownloadHelm",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(Download),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		downloadBinaries,
	}
}

type GetMinikubeProfile struct {
	common.KubeAction
}

func (t *GetMinikubeProfile) Execute(runtime connector.Runtime) error {
	var minikubecmd, ok = t.PipelineCache.GetMustString(common.CacheCommandMinikubePath)
	if !ok || minikubecmd == "" {
		minikubecmd = path.Join(common.BinDir, "minikube")
	}
	var cmd = fmt.Sprintf("%s -p %s profile list -o json --light=false", minikubecmd, runtime.GetRunner().Host.GetMinikubeProfile())
	stdout, err := runtime.GetRunner().Host.CmdExt(cmd, false, false)
	if err != nil {
		return err
	}

	var p mk.Minikube
	if err := json.Unmarshal([]byte(stdout), &p); err != nil {
		return err
	}

	if p.Valid == nil || len(p.Valid) == 0 {
		return fmt.Errorf("minikube profile not found")
	}

	var nodeIp string
	for _, v := range p.Valid {
		if v.Name != runtime.GetRunner().Host.GetMinikubeProfile() {
			continue
		}
		if v.Config.Nodes == nil || len(v.Config.Nodes) == 0 {
			return fmt.Errorf("minikube node not found")
		}
		nodeIp = v.Config.Nodes[0].IP
	}

	if nodeIp == "" {
		return fmt.Errorf("minikube node ip is empty")
	}

	if !util.IsExist(common.KubeAddonsDir) {
		if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("mkdir -p %s", common.KubeAddonsDir), false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create dir %s failed", common.KubeAddonsDir))
		}
	}

	t.PipelineCache.Set(common.CacheMinikubeNodeIp, nodeIp)

	return nil

}

type InitMinikubeNs struct {
	common.KubeAction
}

func (t *InitMinikubeNs) Execute(runtime connector.Runtime) error {
	var kubectlcmd, ok = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if !ok || kubectlcmd == "" {
		kubectlcmd = path.Join(common.BinDir, "kubectl")
	}

	var allNs = []string{
		common.NamespaceKubekeySystem,
		common.NamespaceKubesphereSystem,
		common.NamespaceKubesphereMonitoringSystem,
	}

	for _, ns := range allNs {
		if stdout, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("%s create ns %s", kubectlcmd, ns), false, true); err != nil {
			if !strings.Contains(stdout, "already exists") {
				logger.Errorf("create ns %s failed: %v", ns, err)
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create namespace %s failed: %v", ns, err))
			}
		}
	}

	filePath := path.Join(common.TmpDir, common.KubeAddonsDir, "clusterconfigurations.yaml")
	deployKubesphereCmd := fmt.Sprintf("%s apply -f %s --force", kubectlcmd, filePath)
	if _, err := runtime.GetRunner().Host.CmdExt(deployKubesphereCmd, false, true); err != nil {
		return errors.Wrapf(errors.WithStack(err), "deploy %s failed", filePath)
	}

	return nil
}

type CheckMacCommandExists struct {
	common.KubeAction
}

func (t *CheckMacCommandExists) Execute(runtime connector.Runtime) error {
	var err error
	var minikubepath string
	var kubectlpath string
	var helmpath string
	var dockerpath string

	if minikubepath, err = util.GetCommand(common.CommandMinikube); err != nil || minikubepath == "" {
		return fmt.Errorf("minikube not found")
	}

	if kubectlpath, err = util.GetCommand(common.CommandKubectl); err != nil || kubectlpath == "" {
		return fmt.Errorf("kubectl not found")
	}

	if helmpath, err = util.GetCommand(common.CommandHelm); err != nil || helmpath == "" {
		return fmt.Errorf("helm not found")
	}

	if dockerpath, err = util.GetCommand(common.CommandDocker); err != nil || dockerpath == "" {
		return fmt.Errorf("docker not found")
	}

	fmt.Println("helm path:", helmpath)
	fmt.Println("kubectl path:", kubectlpath)
	fmt.Println("minikube path:", minikubepath)
	fmt.Println("docker path:", dockerpath)

	t.PipelineCache.Set(common.CacheCommandHelmPath, helmpath)
	t.PipelineCache.Set(common.CacheCommandMinikubePath, minikubepath)
	t.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)
	t.PipelineCache.Set(common.CacheCommandDockerPath, dockerpath)

	return nil
}

type CheckMacOsCommandModule struct {
	common.KubeModule
}

func (m *CheckMacOsCommandModule) Init() {
	m.Name = "CheckCommandPath"

	checkMacCommandExists := &task.LocalTask{
		Name:   "CheckMiniKubeExists",
		Action: new(CheckMacCommandExists),
	}

	m.Tasks = []task.Interface{
		checkMacCommandExists,
	}
}

type DeployMiniKubeModule struct {
	common.KubeModule
}

func (m *DeployMiniKubeModule) Init() {
	m.Name = "DeployMiniKube"

	getMinikubeProfile := &task.RemoteTask{
		Name:     "GetMinikubeProfile",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetMinikubeProfile),
		Parallel: false,
		Retry:    1,
	}

	generateManifests := &task.RemoteTask{
		Name:  "GenerateKsInstallerCRD",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Action: &action.Template{
			Name:     "GenerateKsInstallerCRD",
			Template: templates.KsInstaller,
			Dst:      filepath.Join(common.KubeAddonsDir, "clusterconfigurations.yaml"),
		},
		Parallel: false,
		Retry:    1,
	}

	initMinikubeNs := &task.LocalTask{
		Name:   "InitMinikubeNs",
		Action: new(InitMinikubeNs),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		getMinikubeProfile,
		generateManifests,
		initMinikubeNs,
	}
}
