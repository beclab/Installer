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

type InstallBfl struct {
	common.KubeAction
}

func (t *InstallBfl) Execute(runtime connector.Runtime) error {
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
	var launchPath = path.Join(runtime.GetInstallerDir(), cc.WizardDir, "wizard", "config", "launcher")
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

	installBfl := &task.LocalTask{
		Name:   "InstallBfl",
		Desc:   "Install Bfl",
		Action: new(InstallBfl),
		Retry:  1,
	}

	checkBflRunning := &task.LocalTask{
		Name: "CheckBflStatus",
		Action: &CheckPodsRunning{
			labels: map[string][]string{
				fmt.Sprintf("user-space-%s", m.KubeConf.Arg.User.UserName): {"tier=bfl"},
			},
		},
		Retry: 20,
		Delay: 10 * time.Second,
	}

	m.Tasks = []task.Interface{
		installBfl,
		checkBflRunning,
	}
}
