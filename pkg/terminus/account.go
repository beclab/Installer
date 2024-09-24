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
	accounttemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

type SetAccountValues struct {
	common.KubeAction
}

func (p *SetAccountValues) Execute(runtime connector.Runtime) error {
	// todo User info
	var accountFile = path.Join(runtime.GetInstallerDir(), "wizard", "config", "account", accounttemplates.AccountValues.Name())
	var data = util.Data{
		"UserName":   p.KubeConf.Arg.User.UserName,
		"Password":   p.KubeConf.Arg.User.Password,
		"Email":      p.KubeConf.Arg.User.Email,
		"DomainName": p.KubeConf.Arg.User.DomainName,
	}

	accountStr, err := util.Render(accounttemplates.AccountValues, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render account template failed")
	}

	if err := util.WriteFile(accountFile, []byte(accountStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write account %s failed", accountFile))
	}

	return nil
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

	var accountPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "account")

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

	// todo  Enter account values

	installAccount := &task.LocalTask{
		Name:   "InstallAccount",
		Action: &InstallAccount{},
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		installAccount,
	}
}
