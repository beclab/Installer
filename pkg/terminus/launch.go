package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/utils"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CheckLauncherState struct {
	common.KubeAction
}

func (t *CheckLauncherState) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	cmd := fmt.Sprintf("%s get pod  -n user-space-${username} -l 'tier=bfl' -o jsonpath='{.items[*].status.phase}'", kubectl)

	launcherphase, _ := runtime.GetRunner().SudoCmdExt(cmd, false, false)
	if launcherphase == "Running" {
		return nil
	}
	return fmt.Errorf("Launcher State is Pending")
}

type InstallLaunch struct {
	common.KubeAction
}

func (t *InstallLaunch) Execute(runtime connector.Runtime) error {
	var ns = fmt.Sprintf("user-space-%s", t.KubeConf.Arg.User.UserName)

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var r = utils.Random()
	var key = fmt.Sprintf("bytetrade_bfl_%d", r)
	var secret, _ = utils.GeneratePassword(16)

	var launchName = fmt.Sprintf("launcher-%s", t.KubeConf.Arg.User.UserName)
	var launchPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "launcher")
	var parms = make(map[string]interface{})
	var sets = make(map[string]interface{})
	sets["bfl.appKey"] = key
	sets["bfl.appSecret"] = secret
	parms["set"] = sets
	parms["force"] = true

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, launchName, launchPath, "", ns, parms, false); err != nil {
		return err
	}

	return nil
}

type InstallLaunchModule struct {
	common.KubeModule
}

func (m *InstallLaunchModule) Init() {
	m.Name = "InstallLauncher"

	installLaunch := &task.RemoteTask{
		Name:     "InstallLauncher",
		Desc:     "Install Launcher",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   new(InstallLaunch),
		Parallel: false,
		Retry:    1,
	}

	checkLauncherState := &task.RemoteTask{
		Name:     "CheckLauncherState",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &CheckLauncherState{},
		Parallel: false,
		Retry:    50,
		Delay:    5 * time.Second,
	}

	m.Tasks = []task.Interface{
		installLaunch,
		checkLauncherState,
	}
}

func getNodeName(kubectl string, userName string, runtime connector.Runtime) (node string, err error) {
	var cmd = fmt.Sprintf("%s get pod  -n user-space-%s -l 'tier=bfl' -o jsonpath='{.items[*].spec.nodeName}'", kubectl, userName)
	if node, err = runtime.GetRunner().Host.CmdExt(cmd, false, false); err != nil {
		return "", err
	}
	if node == "" {
		return "", fmt.Errorf("node not found")
	}

	return
}

func getDocUrl(runtime connector.Runtime) (url string, err error) {
	var nodeip string
	var cmd = fmt.Sprintf(`curl --connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 -s http://checkip.dyndns.org/ | grep -o "[[:digit:].]\+"`)
	nodeip, _ = runtime.GetRunner().Host.CmdExt(cmd, false, false)
	url = fmt.Sprintf("http://%s:30883/bfl/apidocs.json", nodeip)
	return
}

func getAnnotation(kubectl, namespace, resType, resName, key string, runtime connector.Runtime) string {
	var cmd = fmt.Sprintf("%s get %s %s -n %s -o jsonpath='{.metadata.annotations.%s}'", kubectl, resType, resName, namespace, key)
	stdout, _ := runtime.GetRunner().Host.CmdExt(cmd, false, true)
	return stdout
}
