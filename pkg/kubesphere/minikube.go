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

// ~ Download
type Download struct {
	common.KubeAction
}

func (t *Download) Execute(runtime connector.Runtime) error {
	var arch = runtime.GetRunner().Host.GetArch()
	helm := files.NewKubeBinary("helm", arch, kubekeyapiv1alpha2.DefaultHelmVersion, runtime.GetWorkDir())

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

// ~ DownloadMinikubeBinaries
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

// ~ GetMinikubeProfile
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

	t.PipelineCache.Set(common.CacheMinikubeNodeIp, nodeIp)

	return nil

}

// ~ InitMinikubeNs
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

// ~ CheckMacCommandExists
type CheckMacCommandExists struct {
	common.KubeAction
}

func (t *CheckMacCommandExists) Execute(runtime connector.Runtime) error {
	var err error
	var minikubepath string
	var kubectlpath string
	var helmpath string

	if minikubepath, err = util.GetCommand(common.CommandMinikube); err != nil || minikubepath == "" {
		return fmt.Errorf("minikube not found")
	}

	if kubectlpath, err = util.GetCommand(common.CommandKubectl); err != nil || kubectlpath == "" {
		return fmt.Errorf("kubectl not found")
	}

	if helmpath, err = util.GetCommand(common.CommandHelm); err != nil || helmpath == "" {
		return fmt.Errorf("helm not found")
	}

	t.PipelineCache.Set(common.CacheCommandHelmPath, helmpath)
	t.PipelineCache.Set(common.CacheCommandMinikubePath, minikubepath)
	t.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)

	return nil
}

// ~ DeployMiniKubeModule
type DeployMiniKubeModule struct {
	common.KubeModule
}

func (m *DeployMiniKubeModule) Init() {
	m.Name = "DeployMiniKube"

	checkMacCommandExists := &task.LocalTask{
		Name:   "CheckMiniKubeExists",
		Action: new(CheckMacCommandExists),
	}

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
		checkMacCommandExists,
		getMinikubeProfile,
		generateManifests,
		initMinikubeNs,
	}
}
