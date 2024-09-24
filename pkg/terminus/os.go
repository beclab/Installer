package terminus

import (
	"context"
	"fmt"
	"path"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
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

	var cmd = fmt.Sprintf("%s get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d", kubectl)
	redisPwd, err := runtime.GetRunner().Host.Cmd(cmd, false, false)
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
	// todo need to fix
	var gpuType = getGpuType(t.KubeConf.Arg.GPU.Enable, t.KubeConf.Arg.GPU.Share)
	var storageBackupBucket = t.KubeConf.Arg.Storage.BackupClusterBucket
	var storageBucket = t.KubeConf.Arg.Storage.StorageBucket
	var storageSyncSecret = t.KubeConf.Arg.Storage.StorageSyncSecret
	var storagePrefix = t.KubeConf.Arg.Storage.StoragePrefix
	var cloudValue = cloudValue(t.KubeConf.Arg.IsCloudInstance)
	var fsType = getFsType(t.KubeConf.Arg.WSL)

	var parms = make(map[string]interface{})
	var sets = make(map[string]interface{})
	sets["kubesphere.redis_password"] = redisPwd
	sets["backup.bucket"] = storageBackupBucket
	sets["backup.key_prefix"] = storagePrefix
	sets["backup.is_cloud_version"] = cloudValue
	sets["backup.sync_secret"] = storageSyncSecret
	sets["gpu"] = gpuType
	sets["s3_bucket"] = storageBucket
	sets["fs_type"] = fsType
	parms["force"] = true
	parms["set"] = sets

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, systemPath, "", common.NamespaceOsSystem, parms, false); err != nil {
		return err
	}

	return nil
}

type InstallOsModule struct {
	common.KubeModule
}

func (m *InstallOsModule) Init() {
	m.Name = "InstallOsSystem"

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
		Delay:  1 * time.Second,
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
	if !gpuEnable {
		return
	} else {
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
