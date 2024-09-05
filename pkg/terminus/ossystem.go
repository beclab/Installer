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

type InstallOsSystemPrepare struct {
	common.KubePrepare
}

func (p *InstallOsSystemPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	kubectlpath, _ := p.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if kubectlpath == "" {
		kubectlpath, err := util.GetCommand(common.CommandKubectl)
		if err != nil {
			return false, fmt.Errorf("kubectl not found")
		}

		p.PipelineCache.Set(common.CacheCommandKubectlPath, kubectlpath)
	}

	redisPwd, _ := p.PipelineCache.GetMustString(common.CacheRedisPassword)
	if redisPwd == "" {
		var cmd = fmt.Sprintf("%s get secret -n kubesphere-system redis-secret -o jsonpath='{.data.auth}' |base64 -d", kubectlpath)
		stdout, err := runtime.GetRunner().Host.CmdExt(cmd, false, false)
		if err != nil {
			return false, err
		}
		if stdout == "" {
			return false, fmt.Errorf("redis secret not exists")
		}
		p.PipelineCache.Set(common.CacheRedisPassword, stdout)
	}

	return true, nil
}

type InstallOsSystem struct {
	common.KubeAction
}

func (t *InstallOsSystem) Execute(runtime connector.Runtime) error {
	var redisPassword, _ = t.PipelineCache.GetMustString(common.CacheRedisPassword)

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

	// var osPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "system")
	var osPath = path.Join(runtime.GetBaseDir(), cc.PackageCacheDir, cc.WizardDir, "wizard", "config", "system")
	var gpuType = getGpuType(t.KubeConf.Arg.GPU.Enable, t.KubeConf.Arg.GPU.Share)
	var storageDomain = getBucket(t.KubeConf.Arg.Storage.StorageDomain) // s3_bucket=${S3_BUCKET}
	var storageBucket = t.KubeConf.Arg.Storage.StorageBucket
	var storageSyncSecret = t.KubeConf.Arg.Storage.StorageSyncSecret
	var storagePrefix = t.KubeConf.Arg.Storage.StoragePrefix
	var cloudValue = cloudValue(t.KubeConf.Arg.IsCloudInstance)
	var fsType = getFsType(t.KubeConf.Arg.WSL)

	var parms = make(map[string]interface{})
	var sets = make(map[string]interface{})
	sets["kubesphere.redis_password"] = redisPassword
	sets["backup.bucket"] = storageBucket
	sets["backup.key_prefix"] = storagePrefix
	sets["backup.is_cloud_version"] = cloudValue
	sets["backup.sync_secret"] = storageSyncSecret
	sets["gpu"] = gpuType
	sets["s3_bucket"] = storageDomain
	sets["fs_type"] = fsType
	parms["force"] = true
	parms["set"] = sets

	if err := utils.UpgradeCharts(ctx, actionConfig, settings, common.ChartNameAccount, osPath, "", common.NamespaceOsSystem, parms, false); err != nil {
		return err
	}

	return nil
}

type InstallOsModule struct {
	common.KubeModule
}

func (m *InstallOsModule) Init() {
	m.Name = "InstallOsSystem"

	installOsSystem := &task.RemoteTask{
		Name:     "InstallOsSystem",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &InstallOsSystem{},
		Parallel: false,
		Retry:    1,
	}

	checkAppServiceState := &task.RemoteTask{
		Name:     "CheckAppServiceState",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &CheckAppServiceState{},
		Parallel: false,
		Retry:    50,
		Delay:    5 * time.Second,
	}

	checkCitusState := &task.RemoteTask{
		Name:     "CheckCitusState",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &CheckCitusState{},
		Parallel: false,
		Retry:    50,
		Delay:    5 * time.Second,
	}

	m.Tasks = []task.Interface{
		installOsSystem,
		checkAppServiceState,
		checkCitusState,
	}
}

func getBucket(storageBucket string) (bucket string) {
	bucket = "none"
	if storageBucket != "" {
		bucket = storageBucket
	}
	return
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
