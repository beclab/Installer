package storage

import (
	"fmt"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	corecommon "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	juicefsTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

// - InstallJuiceFsModule
type InstallJuiceFsModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *InstallJuiceFsModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallJuiceFsModule) Init() {
	m.Name = "InstallJuiceFs"

	installJuiceFs := &task.RemoteTask{
		Name:    "InstallJuiceFs",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckJuiceFsExists{},
		Action: &InstallJuiceFs{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	configJuiceFsMetaDB := &task.RemoteTask{
		Name:     "ConfigJuiceFSMetaDB",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(ConfigJuiceFsMetaDB),
		Parallel: false,
		Retry:    1,
	}

	enableJuiceFsService := &task.RemoteTask{
		Name:     "EnableJuiceFsService",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(EnableJuiceFsService),
		Parallel: false,
		Retry:    1,
	}

	checkJuiceFsState := &task.RemoteTask{
		Name:     "CheckJuiceFsState",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(CheckJuiceFsState),
		Parallel: false,
		Retry:    5,
		Delay:    5 * time.Second,
	}

	m.Tasks = []task.Interface{
		installJuiceFs,
		configJuiceFsMetaDB,
		enableJuiceFsService,
		checkJuiceFsState,
	}
}

type CheckJuiceFsExists struct {
	common.KubePrepare
}

func (p *CheckJuiceFsExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if utils.IsExist(JuiceFsFile) {
		return false, nil
	}

	return true, nil
}

type InstallJuiceFs struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallJuiceFs) Execute(runtime connector.Runtime) error {
	if !utils.IsExist(JuiceFsDataDir) {
		err := utils.Mkdir(JuiceFsDataDir)
		if err != nil {
			return err
		}
	}

	juicefs, err := t.Manifest.Get("juicefs")
	if err != nil {
		return err
	}

	path := juicefs.FilePath(t.BaseDir)

	var cmd = fmt.Sprintf("rm -rf /tmp/juicefs* && cp -f %s /tmp/%s && cd /tmp && tar -zxf ./%s && chmod +x juicefs && install juicefs /usr/local/bin && install juicefs /sbin/mount.juicefs && rm -rf ./LICENSE ./README.md ./README_CN.md ./juicefs*", path, juicefs.Filename, juicefs.Filename)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

type EnableJuiceFsService struct {
	common.KubeAction
}

func (t *EnableJuiceFsService) Execute(runtime connector.Runtime) error {
	localIP := runtime.GetSystemInfo().GetLocalIp()
	var redisPassword, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var redisService = fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, localIP)
	var data = util.Data{
		"JuiceFsBinPath":    JuiceFsFile,
		"JuiceFsCachePath":  JuiceFsCacheDir,
		"JuiceFsMetaDb":     redisService,
		"JuiceFsMountPoint": OlaresJuiceFSRootDir,
	}

	juiceFsServiceStr, err := util.Render(juicefsTemplates.JuicefsService, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render juicefs service template failed")
	}
	if err := util.WriteFile(JuiceFsServiceFile, []byte(juiceFsServiceStr), corecommon.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write juicefs service %s failed", JuiceFsServiceFile))
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl restart juicefs", false, false); err != nil {
		return err
	}

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl enable juicefs", false, false); err != nil {
		return err
	}

	return nil
}

type ConfigJuiceFsMetaDB struct {
	common.KubeAction
}

func (t *ConfigJuiceFsMetaDB) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	var redisPassword, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var redisService = fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, localIp)
	if redisPassword == "" {
		return fmt.Errorf("redis password not found")
	}

	storageFlags, err := getStorageFlags(t.KubeConf.Arg.Storage, localIp)
	if err != nil {
		return err
	}
	var cmd = fmt.Sprintf("%s format %s", JuiceFsFile, redisService)
	cmd = cmd + storageFlags

	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}
	return nil
}

type CheckJuiceFsState struct {
	common.KubeAction
}

func (t *CheckJuiceFsState) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl --no-pager -n 0 status juicefs", false, false); err != nil {
		return fmt.Errorf("JuiceFs Pending")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s summary %s", JuiceFsFile, OlaresJuiceFSRootDir), false, false); err != nil {
		return err
	}

	return nil
}

func getStorageFlags(storage *common.Storage, localIp string) (string, error) {
	var storageFlags string
	var fsName string
	var err error

	switch storage.StorageType {
	case common.ManagedMinIO:
		storageFlags, err = getManagedMinIOAccessFlags(localIp)
		if err != nil {
			return "", err
		}
	default:
		storageFlags = getExternalStorageAccessFlags(storage)
	}

	if storage.StorageVendor == "true" && storage.StorageClusterId != "" {
		fsName = storage.StorageClusterId
	} else {
		fsName = "rootfs"
	}

	storageFlags = storageFlags + fmt.Sprintf(" %s --trash-days 0", fsName)

	return storageFlags, nil
}

func getExternalStorageAccessFlags(storage *common.Storage) string {
	var params = fmt.Sprintf(" --storage %s --bucket %s", storage.StorageType, storage.StorageBucket)
	if storage.StorageVendor == "true" {
		if storage.StorageToken != "" {
			params = params + fmt.Sprintf(" --session-token %s", storage.StorageToken)
		}
	}
	if storage.StorageAccessKey != "" {
		params = params + fmt.Sprintf(" --access-key %s", storage.StorageAccessKey)
	}
	if storage.StorageSecretKey != "" {
		params = params + fmt.Sprintf(" --secret-key %s", storage.StorageSecretKey)
	}

	return params
}

func getManagedMinIOAccessFlags(localIp string) (string, error) {
	minioPassword, err := getMinioPwdFromConfigFile()
	if err != nil {
		return "", errors.Wrap(err, "failed to get password of managed MinIO")
	}
	return fmt.Sprintf(" --storage minio --bucket http://%s:9000/%s --access-key %s --secret-key %s",
		localIp, cc.OlaresDir, MinioRootUser, minioPassword), nil
}
