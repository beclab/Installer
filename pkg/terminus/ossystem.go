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
	configmaptemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

type CheckAppServiceState struct {
	common.KubeAction
}

func (t *CheckAppServiceState) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	cmd := fmt.Sprintf("%s get pod  -n os-system -l 'tier=app-service' -o jsonpath='{.items[*].status.phase}'", kubectl)

	appservicephase, _ := runtime.GetRunner().SudoCmdExt(cmd, false, false)
	if appservicephase == "Running" {
		return nil
	}
	return fmt.Errorf("App Service State is Pending")
}

type CheckCitusState struct {
	common.KubeAction
}

func (t *CheckCitusState) Execute(runtime connector.Runtime) error {
	kubectl, _ := t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	cmd := fmt.Sprintf("%s get pod  -n os-system -l 'app=citus' -o jsonpath='{.items[*].status.phase}'", kubectl)

	citusphase, _ := runtime.GetRunner().SudoCmdExt(cmd, false, false)
	if citusphase == "Running" {
		return nil
	}
	return fmt.Errorf("Citus State is Pending")
}

type InstallOsSystem struct {
	common.KubeAction
}

func (t *InstallOsSystem) Execute(runtime connector.Runtime) error {
	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}

	var sharedLib = "/terminus/share"
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("mkdir -p %s && chown 1000:1000 %s", sharedLib, sharedLib), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), "create /terminus/share failed")
	}

	var cmd = fmt.Sprintf("%s get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d", kubectl)
	redisPwd, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "get redis secret error")
	}

	if redisPwd == "" {
		return fmt.Errorf("redis secret not found")
	}

	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, common.NamespaceOsSystem)
	if err != nil {
		return err
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var systemPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "system")
	var parms = make(map[string]interface{})
	var sets = make(map[string]interface{})
	sets["kubesphere.redis_password"] = redisPwd
	sets["backup.bucket"] = t.KubeConf.Arg.Storage.BackupClusterBucket
	sets["backup.key_prefix"] = t.KubeConf.Arg.Storage.StoragePrefix
	sets["backup.is_cloud_version"] = cloudValue(t.KubeConf.Arg.IsCloudInstance)
	sets["backup.sync_secret"] = t.KubeConf.Arg.Storage.StorageSyncSecret
	sets["gpu"] = getGpuType(t.KubeConf.Arg.GPU.Enable, t.KubeConf.Arg.GPU.Share)
	sets["s3_bucket"] = t.KubeConf.Arg.Storage.StorageBucket
	sets["fs_type"] = getFsType(t.KubeConf.Arg.WSL)
	sets["sharedlib"] = sharedLib
	parms["force"] = true
	parms["set"] = sets

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, systemPath, "", common.NamespaceOsSystem, parms, false); err != nil {
		return err
	}

	return nil
}

type CreateBackupConfigMap struct {
	common.KubeAction
}

func (t *CreateBackupConfigMap) Execute(runtime connector.Runtime) error {
	var backupConfigMapFile = path.Join(runtime.GetInstallerDir(), "deploy", configmaptemplates.BackupConfigMap.Name())
	var data = util.Data{
		"CloudInstance":     cloudValue(t.KubeConf.Arg.IsCloudInstance),
		"StorageBucket":     t.KubeConf.Arg.Storage.BackupClusterBucket,
		"StoragePrefix":     t.KubeConf.Arg.Storage.StoragePrefix,
		"StorageSyncSecret": t.KubeConf.Arg.Storage.StorageSyncSecret,
	}

	backupConfigStr, err := util.Render(configmaptemplates.BackupConfigMap, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render backup configmap template failed")
	}
	if err := util.WriteFile(backupConfigMapFile, []byte(backupConfigStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write backup configmap %s failed", backupConfigMapFile))
	}

	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, backupConfigMapFile), false, true); err != nil {
		return err
	}

	return nil
}

type CreateReverseProxyConfigMap struct {
	common.KubeAction
}

func (c *CreateReverseProxyConfigMap) Execute(runtime connector.Runtime) error {
	var defaultReverseProxyConfigMapFile = path.Join(runtime.GetInstallerDir(), "deploy", configmaptemplates.ReverseProxyConfigMap.Name())
	var data = util.Data{
		"EnableCloudflare": c.KubeConf.Arg.Cloudflare.Enable,
		"EnableFrp":        c.KubeConf.Arg.Frp.Enable,
		"FrpServer":        c.KubeConf.Arg.Frp.Server,
		"FrpPort":          c.KubeConf.Arg.Frp.Port,
		"FrpAuthMethod":    c.KubeConf.Arg.Frp.AuthMethod,
		"FrpAuthToken":     c.KubeConf.Arg.Frp.AuthToken,
	}

	reverseProxyConfigStr, err := util.Render(configmaptemplates.ReverseProxyConfigMap, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render backup configmap template failed")
	}
	if err := util.WriteFile(defaultReverseProxyConfigMapFile, []byte(reverseProxyConfigStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write backup configmap %s failed", defaultReverseProxyConfigMapFile))
	}

	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, defaultReverseProxyConfigMapFile), false, true); err != nil {
		return err
	}

	return nil
}

type Patch struct {
	common.KubeAction
}

func (p *Patch) Execute(runtime connector.Runtime) error {
	var kubectl, _ = util.GetCommand(common.CommandKubectl)
	var globalRoleWorkspaceManager = path.Join(runtime.GetInstallerDir(), "deploy", "patch-globalrole-workspace-manager.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, globalRoleWorkspaceManager), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "patch globalrole workspace manager failed")
	}

	var notificationManager = path.Join(runtime.GetInstallerDir(), "deploy", "patch-notification-manager.yaml")
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s apply -f %s", kubectl, notificationManager), false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "patch notification manager failed")
	}

	return nil
}

type InstallOsSystemModule struct {
	common.KubeModule
}

func (m *InstallOsSystemModule) Init() {
	m.Name = "InstallOsSystemModule"

	installOsSystem := &task.LocalTask{
		Name:   "InstallOsSystem",
		Action: &InstallOsSystem{},
		Retry:  1,
	}

	// todo cm-backup-config
	// todo patchs

	checkAppServiceState := &task.LocalTask{
		Name:   "CheckAppServiceState",
		Action: &CheckAppServiceState{},
		Retry:  20,
		Delay:  10 * time.Second,
	}

	checkCitusState := &task.LocalTask{
		Name:   "CheckCitusState",
		Action: &CheckCitusState{},
		Retry:  20,
		Delay:  10 * time.Second,
	}

	m.Tasks = []task.Interface{
		installOsSystem,
		checkAppServiceState,
		checkCitusState,
	}
}

func getGpuType(gpuEnable, gpuShare bool) (gpuType string) {
	gpuType = "none"
	if gpuEnable {
		if gpuShare {
			gpuType = "nvshare"
		} else {
			gpuType = "nvidia"
		}
	}

	return gpuType
}

func cloudValue(cloudInstance bool) string {
	if cloudInstance {
		return "true"
	}

	return ""
}

func getFsType(wsl bool) string {
	if wsl {
		return "fs"
	}
	return "jfs"
}
