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
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"

	ctrl "sigs.k8s.io/controller-runtime"
)

type InstallAccount struct {
	common.KubeAction
}

func (t *InstallAccount) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, "")
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var accountPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "account")
	if !util.IsExist(accountPath) {
		return fmt.Errorf("account not exists")
	}

	var args = make(map[string]interface{})
	args["force"] = true
	if t.KubeConf.Arg.WSL {
		var sets = make(map[string]interface{})
		sets["nat_gateway_ip"] = "" // todo natgateway
		args["set"] = sets
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, accountPath, "", "", args, false); err != nil {
		return err
	}

	return nil
}

type InstallAccountModule struct {
	common.KubeModule
}

func (m *InstallAccountModule) Init() {
	m.Name = "InstallAccount"

	installAccount := &task.RemoteTask{
		Name:     "InstallAccount",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &InstallAccount{},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installAccount,
	}
}
