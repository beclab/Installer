package storage

import (
	"fmt"
	"os/exec"
	"time"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"github.com/pkg/errors"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	minioTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"bytetrade.io/web3os/installer/pkg/utils"
)

// ~ CheckMinioState
type CheckMinioState struct {
	common.KubeAction
}

func (t *CheckMinioState) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl --no-pager -n 0 status minio" // 这里可以考虑用 is-active 来验证
	stdout, err := runtime.GetRunner().SudoCmdExt(cmd, false, false)
	if err != nil {
		return fmt.Errorf("Minio Pending")
	}

	logger.Debug(stdout)

	return nil
}

// ~ EnableMinio
type EnableMinio struct {
	common.KubeAction
}

func (t *EnableMinio) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmdExt("groupadd -r minio", false, false)
	_, _ = runtime.GetRunner().SudoCmdExt("useradd -M -r -g minio minio", false, false)
	_, _ = runtime.GetRunner().SudoCmdExt(fmt.Sprintf("chown minio:minio %s", MinioDataDir), false, false)

	if _, err := runtime.GetRunner().SudoCmdExt("systemctl daemon-reload", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl restart minio", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmd("systemctl enable minio", false, false); err != nil {
		return err
	}

	return nil
}

// ~ ConfigMinio
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

// ~ InstallMinio
type InstallMinio struct {
	common.KubeAction
}

func (t *InstallMinio) Execute(runtime connector.Runtime) error {
	var arch = constants.OsArch
	binary := files.NewKubeBinary("minio", arch, kubekeyapiv1alpha2.DefaultMinioVersion, runtime.GetWorkDir())

	if err := binary.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
	}

	var exists = util.IsExist(binary.Path())
	if exists {
		p := binary.Path()
		if err := binary.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		}
		// _ = exec.Command(fmt.Sprintf("cp %s /usr/local/bin/", binary.Path())).Run()
		// return nil
	}

	if !exists || binary.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)
		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	var cmd = fmt.Sprintf("cd %s && chmod +x minio && install minio /usr/local/bin", binary.BaseDir)
	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
		return err
	}

	t.PipelineCache.Set(common.CacheMinioPath, binary.Path())

	return nil
}

// ~ CheckMinioExists
type CheckMinioExists struct {
	common.KubePrepare
}

func (p *CheckMinioExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if !utils.IsExist(MinioDataDir) {
		utils.Mkdir(MinioDataDir)
	}

	if !utils.IsExist(MinioFile) {
		return true, nil
	}

	return false, nil
}

// - InstallMinioModule
type InstallMinioModule struct {
	common.KubeModule
	Skip bool
}

func (m *InstallMinioModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallMinioModule) Init() {
	m.Name = "InstallMinio"

	installMinio := &task.RemoteTask{
		Name:     "InstallMinio",
		Hosts:    m.Runtime.GetAllHosts(),
		Prepare:  &CheckMinioExists{},
		Action:   &InstallMinio{},
		Parallel: false,
	}

	configMinio := &task.RemoteTask{
		Name:     "ConfigMinio",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   &ConfigMinio{},
		Parallel: false,
	}

	enableMinio := &task.RemoteTask{
		Name:     "EnableMinioService",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   &EnableMinio{},
		Parallel: false,
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
