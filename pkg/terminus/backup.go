package terminus

import (
	"fmt"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	configmaptemplates "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"github.com/pkg/errors"
)

type CreateBackupConfigMap struct {
	common.KubeAction
}

func (t *CreateBackupConfigMap) Execute(runtime connector.Runtime) error {
	var installPath = filepath.Dir(t.KubeConf.Arg.Manifest)
	var backupConfigMapFile = path.Join(installPath, "deploy", configmaptemplates.BackupConfigMap.Name())
	var data = util.Data{
		"CloudInstance":     cloudValue(t.KubeConf.Arg.IsCloudInstance),
		"StorageBucket":     t.KubeConf.Arg.Storage.StorageBucket,
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

	var kubectlpath, _ = t.PipelineCache.GetMustString(common.CacheCommandKubectlPath)
	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("%s apply -f %s", kubectlpath, backupConfigMapFile), false, true); err != nil {
		return err
	}

	return nil
}

type CreateBackupConfigMapModule struct {
	common.KubeModule
}

func (m *CreateBackupConfigMapModule) Init() {
	m.Name = "CreateBackupConfigMap"

	createBackupConfigMap := &task.RemoteTask{
		Name:     "CreateBackupConfigMap",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.IsMaster),
		Action:   &CreateBackupConfigMap{},
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		createBackupConfigMap,
	}
}
