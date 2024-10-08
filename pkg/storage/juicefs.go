package storage

import (
	"fmt"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/cache"
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
	// todo redis password fetch
	// var redisPassword, err = getRedisPwd(runtime)
	// if err != nil {
	// 	return err
	// }

	// minioPassword, err := getMinioPwd(runtime)
	// if err != nil {
	// 	return err
	// }

	var redisPassword, ok = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)

	if !ok || redisPassword == "" {
		return fmt.Errorf("redis password not found")
	}

	var storageStr = getStorageTypeStr(t.PipelineCache, t.KubeConf.Arg.Storage)

	var redisService = fmt.Sprintf("redis://:%s@%s:6379/1", redisPassword, constants.LocalIp)
	var cmd = fmt.Sprintf("%s format %s --storage %s", JuiceFsFile, redisService, t.KubeConf.Arg.Storage.StorageType)
	cmd = cmd + storageStr

	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return err
	}

	var data = util.Data{
		"JuiceFsBinPath":    JuiceFsFile,
		"JuiceFsCachePath":  JuiceFsCacheDir,
		"JuiceFsMetaDb":     redisService,
		"JuiceFsMountPoint": JuiceFsMountPointDir,
	}

	juiceFsServiceStr, err := util.Render(juicefsTemplates.JuicefsService, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render juicefs service template failed")
	}
	if err := util.WriteFile(JuiceFsServiceFile, []byte(juiceFsServiceStr), corecommon.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write juicefs service %s failed", JuiceFsServiceFile))
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		return errors.Wrap(errors.WithStack(err), "run juicefs failed")
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

func getRedisPwd(runtime connector.Runtime) (string, error) {
	var cmd = fmt.Sprintf("cat /terminus/data/redis/etc/redis.conf 2>&1 |grep requirepass |cut -d' ' -f2 |tr -d '\n'")
	if res, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get redis password error")
	} else if res == "" {
		return "", fmt.Errorf("redis password not found")
	} else {
		return res, nil
	}
}

func getMinioPwd(runtime connector.Runtime) (string, error) {
	var cmd = fmt.Sprintf("cat /etc/default/minio 2>&1 |grep 'MINIO_ROOT_PASSWORD=' |cut -d'=' -f2 |tr -d '\n'")
	if res, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get minio password error")
	} else if res == "" {
		return "", fmt.Errorf("minio password not found")
	} else {
		return res, nil
	}
}

type CheckJuiceFsState struct {
	common.KubeAction
}

func (t *CheckJuiceFsState) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl --no-pager -n 0 status juicefs", false, false); err != nil {
		return fmt.Errorf("JuiceFs Pending")
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("test -d %s/.trash", JuiceFsMountPointDir), false, false); err != nil {
		return err
	}

	return nil
}

func getStorageTypeStr(pc *cache.Cache, storage *common.Storage) string {
	var storageType = storage.StorageType
	var formatStr string
	var fsName string

	switch storageType {
	case common.Minio:
		formatStr = getMinioStr(pc)
	case common.OSS, common.S3:
		formatStr = getCloudStr(storage)
	}

	if storage.StorageVendor == "true" {
		fsName = storage.StorageClusterId
	} else {
		fsName = "rootfs"
	}

	formatStr = formatStr + fmt.Sprintf(" %s --trash-days 0", fsName)

	return formatStr
}

func getCloudStr(storage *common.Storage) string {

	var str = fmt.Sprintf(" --bucket %s", storage.StorageBucket)
	if storage.StorageVendor == "true" {
		if storage.StorageToken != "" {
			str = str + fmt.Sprintf(" --session-token %s", storage.StorageToken)
		}
	}
	if storage.StorageAccessKey != "" && storage.StorageSecretKey != "" {
		str = str + fmt.Sprintf(" --access-key %s --secret-key %s", storage.StorageAccessKey, storage.StorageSecretKey)
	}

	return str
}

func getMinioStr(pc *cache.Cache) string {
	var minioPassword, _ = pc.GetMustString(common.CacheMinioPassword)
	return fmt.Sprintf(" --bucket http://%s:9000/%s --access-key %s --secret-key %s",
		constants.LocalIp, cc.TerminusDir, MinioRootUser, minioPassword)

	// return fmt.Sprintf(" --bucket http://%s:9000/%s --access-key %s --secret-key %s",
	// 	constants.LocalIp, cc.TerminusDir, MinioRootUser, minioPassword)
}
