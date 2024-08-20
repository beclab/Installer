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

type InstallSettings struct {
	common.KubeAction
}

func (t *InstallSettings) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, "")
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	var settingsPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "settings")
	if !util.IsExist(settingsPath) {
		return fmt.Errorf("settings not exists")
	}
	var args = make(map[string]interface{})
	args["force"] = true

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameSettings, settingsPath, "", "", args, false); err != nil {
		return err
	}

	return nil
}

type InstallSettingsModule struct {
	common.KubeModule
}

func (m *InstallSettingsModule) Init() {
	m.Name = "InstallSettings"

	installSettings := &task.RemoteTask{
		Name:     "InstallAccount",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &InstallSettings{},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installSettings,
	}
}
