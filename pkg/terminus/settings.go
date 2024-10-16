package terminus

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	settingstemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type UpdateSettingsValuePrepare struct {
	common.KubePrepare
}

func (p *UpdateSettingsValuePrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	var installPath = filepath.Dir(p.KubeConf.Arg.Manifest)
	var settingsFile = path.Join(installPath, "wizard", "config", "settings", settingstemplates.SettingsValue.Name())
	var data = util.Data{}

	settingsStr, err := util.Render(settingstemplates.SettingsValue, data)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "render settings template failed")
	}

	if err := util.WriteFile(settingsFile, []byte(settingsStr), cc.FileMode0644); err != nil {
		return false, errors.Wrap(errors.WithStack(err), fmt.Sprintf("write settings %s failed", settingsFile))
	}

	return true, nil
}

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

	var installPath = filepath.Dir(t.KubeConf.Arg.Manifest)
	var settingsPath = path.Join(runtime.GetBaseDir(), installPath, "wizard", "config", "settings")
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
