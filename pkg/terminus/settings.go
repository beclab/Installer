package terminus

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"os"
	"path"
	"strings"
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

type SetSettingsValues struct {
	common.KubeAction
}

func (p *SetSettingsValues) Execute(runtime connector.Runtime) error {
	s3SessionToken := "none"
	if p.KubeConf.Arg.Storage.StorageToken != "" {
		s3SessionToken = p.KubeConf.Arg.Storage.StorageToken
	}
	s3AccessKey := "none"
	if p.KubeConf.Arg.Storage.StorageAccessKey != "" {
		s3AccessKey = p.KubeConf.Arg.Storage.StorageAccessKey
	}
	s3SecretKey := "none"
	if p.KubeConf.Arg.Storage.StorageSecretKey != "" {
		s3SecretKey = p.KubeConf.Arg.Storage.StorageSecretKey
	}

	terminusdInstalled := "0"
	if !runtime.GetSystemInfo().IsDarwin() {
		terminusdInstalled = "1"
	}

	var settingsFile = path.Join(runtime.GetInstallerDir(), "wizard", "config", "settings", settingstemplates.SettingsValue.Name())
	var data = util.Data{
		"UserName":           p.KubeConf.Arg.User.UserName,
		"S3SessionToken":     s3SessionToken,
		"S3AccessKeyId":      s3AccessKey,
		"S3SecretAccessKey":  s3SecretKey,
		"ClusterID":          p.KubeConf.Arg.Storage.StorageClusterId,
		"TerminusdInstalled": terminusdInstalled,
	}

	settingsStr, err := util.Render(settingstemplates.SettingsValue, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render settings template failed")
	}

	if err := util.WriteFile(settingsFile, []byte(settingsStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write settings %s failed", settingsFile))
	}

	return nil
}

type MutateTerminusCRFile struct {
	common.KubeAction
}

func (p *MutateTerminusCRFile) Execute(runtime connector.Runtime) error {
	terminusCRFile := path.Join(runtime.GetInstallerDir(), "wizard/config/settings/templates/terminus_cr.yaml")
	byteContent, err := os.ReadFile(terminusCRFile)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "read terminus cr file failed")
	}
	content := string(byteContent)
	content = strings.ReplaceAll(content, "#__DOMAIN_NAME__", p.KubeConf.Arg.User.DomainName)
	selfhosted := "true"
	if p.KubeConf.Arg.IsCloudInstance {
		selfhosted = "false"
	}
	if strings.Contains(p.KubeConf.Arg.PublicNetworkInfo.Hostname, common.CloudVendorAWS) && !strings.Contains(p.KubeConf.Arg.PublicNetworkInfo.PublicIp, "Not Found") {
		selfhosted = "false"
	}
	content = strings.ReplaceAll(content, "#__SELFHOSTED__", selfhosted)
	if err := util.WriteFile(terminusCRFile, []byte(content), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), "write terminus cr file failed")
	}
	return nil
}

type InstallSettings struct {
	common.KubeAction
}

func (t *InstallSettings) Execute(runtime connector.Runtime) error {
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	ns := corev1.NamespaceDefault
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	var settingsPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "settings")
	if !util.IsExist(settingsPath) {
		return fmt.Errorf("settings not exists")
	}

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameSettings, settingsPath, "", ns, nil, false); err != nil {
		return err
	}

	return nil
}

type InstallSettingsModule struct {
	common.KubeModule
}

func (m *InstallSettingsModule) Init() {
	logger.InfoInstallationProgress("Installing settings ...")
	m.Name = "InstallSettings"

	setSettingsValues := &task.LocalTask{
		Name:   "SetSettingsValues",
		Action: new(SetSettingsValues),
		Retry:  1,
	}

	mutateTerminusCRFile := &task.LocalTask{
		Name:    "MutateTerminusCRFile",
		Prepare: new(CheckPublicNetworkInfo),
		Action:  new(MutateTerminusCRFile),
		Retry:   1,
	}

	installSettings := &task.LocalTask{
		Name:   "InstallSettings",
		Action: new(InstallSettings),
		Retry:  1,
	}

	m.Tasks = []task.Interface{
		setSettingsValues,
		mutateTerminusCRFile,
		installSettings,
	}
}
