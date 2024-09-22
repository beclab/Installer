package storage

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	minioTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type CheckMinioState struct {
	common.KubeAction
}

func (t *CheckMinioState) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl --no-pager -n 0 status minio" //
	stdout, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	if err != nil {
		return fmt.Errorf("Minio Pending")
	}

	logger.Debug(stdout)

	return nil
}

type EnableMinio struct {
	common.KubeAction
}

func (t *EnableMinio) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().Host.SudoCmd("groupadd -r minio", false, false)
	_, _ = runtime.GetRunner().Host.SudoCmd("useradd -M -r -g minio minio", false, false)
	_, _ = runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("chown minio:minio %s", MinioDataDir), false, false)

	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl restart minio", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl enable minio", false, false); err != nil {
		return err
	}

	return nil
}

type ConfigMinio struct {
	common.KubeAction
}

func (t *ConfigMinio) Execute(runtime connector.Runtime) error {
	// write file
	var minioPassword, _ = utils.GeneratePassword(16)
	var data = util.Data{
		"MinioCommand": MinioFile,
	}
	minioServiceStr, err := util.Render(minioTemplates.MinioService, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render minio service template failed")
	}
	if err := util.WriteFile(MinioServiceFile, []byte(minioServiceStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write minio service %s failed", MinioServiceFile))
	}

	data = util.Data{
		"MinioDataPath": MinioDataDir,
		"LocalIP":       constants.LocalIp,
		"User":          MinioRootUser,
		"Password":      minioPassword,
	}
	minioEnvStr, err := util.Render(minioTemplates.MinioEnv, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render minio env template failed")
	}

	if err := util.WriteFile(MinioConfigFile, []byte(minioEnvStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write minio env %s failed", MinioConfigFile))
	}

	t.PipelineCache.Set(common.CacheMinioPassword, minioPassword)

	return nil
}

type InstallMinio struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallMinio) Execute(runtime connector.Runtime) error {
	if !utils.IsExist(MinioDataDir) {
		err := utils.Mkdir(MinioDataDir)
		if err != nil {
			logger.Errorf("cannot mkdir %s for minio", MinioDataDir)
			return err
		}
	}

	minio, err := t.Manifest.Get("minio")
	if err != nil {
		return err
	}

	path := minio.FilePath(t.BaseDir)

	// var cmd = fmt.Sprintf("cd %s && chmod +x minio && install minio /usr/local/bin", minio.BaseDir)
	var cmd = fmt.Sprintf("cp -f %s /tmp/minio && chmod +x /tmp/minio && install /tmp/minio /usr/local/bin", path)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return err
	}

	return nil
}

type CheckMinioExists struct {
	common.KubePrepare
}

func (p *CheckMinioExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if !utils.IsExist(MinioDataDir) {
		return true, nil
	}

	if !utils.IsExist(MinioFile) {
		return true, nil
	}

	return false, nil
}

// - InstallMinioModule
type InstallMinioModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *InstallMinioModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallMinioModule) Init() {
	m.Name = "InstallMinio"

	installMinio := &task.RemoteTask{
		Name:    "InstallMinio",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckMinioExists{},
		Action: &InstallMinio{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    1,
	}

	configMinio := &task.RemoteTask{
		Name:     "ConfigMinio",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   &ConfigMinio{},
		Parallel: false,
		Retry:    1,
	}

	enableMinio := &task.RemoteTask{
		Name:     "EnableMinioService",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   &EnableMinio{},
		Parallel: false,
		Retry:    1,
	}

	checkMinioState := &task.RemoteTask{
		Name:     "CheckMinioState",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   &CheckMinioState{},
		Parallel: false,
		Retry:    30,
		Delay:    2 * time.Second,
	}

	m.Tasks = []task.Interface{
		installMinio,
		configMinio,
		enableMinio,
		checkMinioState,
	}
}
