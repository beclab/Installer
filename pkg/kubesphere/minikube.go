package kubesphere

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere/templates"
	mk "bytetrade.io/web3os/installer/pkg/version/minikube"
	"github.com/pkg/errors"
)

// ~ GetMinikubeProfile
type GetMinikubeProfile struct {
	common.KubeAction
}

func (t *GetMinikubeProfile) Execute(runtime connector.Runtime) error {
	var cmd = fmt.Sprintf("minikube -p %s profile list -o json --light=false", runtime.GetRunner().Host.GetMinikubeProfile())
	stdout, err := runtime.GetRunner().SudoCmdExt(cmd, false, false)
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
	var allNs = []string{
		common.NamespaceKubekeySystem,
		common.NamespaceKubesphereSystem,
		common.NamespaceKubesphereMonitoringSystem,
	}

	for _, ns := range allNs {
		if stdout, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("/usr/local/bin/kubectl create ns %s", ns), false, true); err != nil {
			if !strings.Contains(stdout, "already exists") {
				logger.Errorf("create ns %s failed: %v", ns, err)
				return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create namespace %s failed: %v", ns, err))
			}
		}
	}

	filePath := path.Join(common.TmpDir, common.KubeAddonsDir, "clusterconfigurations.yaml")
	deployKubesphereCmd := fmt.Sprintf("/usr/local/bin/kubectl apply -f %s --force", filePath)
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
	if m, err := util.GetCommand(common.CommandMinikube); err != nil || m == "" {
		return fmt.Errorf("minikube not found")
	}

	if d, err := util.GetCommand(common.CommandDocker); err != nil || d == "" {
		return fmt.Errorf("docker not found")
	}

	if h, err := util.GetCommand(common.CommandHelm); err != nil || h == "" {
		return fmt.Errorf("helm not found")
	}

	if h, err := util.GetCommand(common.CommandKubectl); err != nil || h == "" {
		return fmt.Errorf("kubectl not found")
	}
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
