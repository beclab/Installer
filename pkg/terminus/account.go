package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	accounttemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

type UpdateAccountValues struct {
	common.KubePrepare
}

func (p *UpdateAccountValues) PreCheck(runtime connector.Runtime) (bool, error) {
	var accountFile = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "account", accounttemplates.AccountValues.Name())
	var data = util.Data{
		"UserName":   p.KubeConf.Arg.User.UserName,
		"Password":   p.KubeConf.Arg.User.Password,
		"Email":      p.KubeConf.Arg.User.Email,
		"DomainName": p.KubeConf.Arg.User.DomainName,
	}

	accountStr, err := util.Render(accounttemplates.AccountValues, data)
	if err != nil {
		return false, errors.Wrap(errors.WithStack(err), "render account template failed")
	}

	if err := util.WriteFile(accountFile, []byte(accountStr), cc.FileMode0644); err != nil {
		return false, errors.Wrap(errors.WithStack(err), fmt.Sprintf("write account %s failed", accountFile))
	}

	return true, nil
}

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
		Name:  "InstallAccount",
		Hosts: m.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.IsMaster),
			new(UpdateAccountValues),
		},
		Action:   &InstallAccount{},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		installAccount,
	}
}
