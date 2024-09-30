package terminus

import (
	"context"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	uninstalltemplate "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"

	"github.com/pkg/errors"
)

type GetTerminusVersion struct {
}

func (t *GetTerminusVersion) Execute() (string, error) {
	var kubectlpath, err = util.GetCommand(common.CommandKubectl)
	if err != nil {
		return "", fmt.Errorf("kubectl not found, Terminus might not be installed.")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", fmt.Sprintf("%s get terminus -o jsonpath='{.items[*].spec.version}'", kubectlpath))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get terminus version failed")
	}

	if version := string(output); version == "" {
		return "", fmt.Errorf("Terminus might not be installed.")
	} else {
		return version, nil
	}
}

type CheckPodsRunning struct {
	common.KubeAction
	labels map[string][]string
}

func (c *CheckPodsRunning) Execute(runtime connector.Runtime) error {
	if c.labels == nil {
		return nil
	}

	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubectl not found")
	}

	var ctx, cancel = context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	for ns, labels := range c.labels {
		for _, label := range labels {
			var cmd = fmt.Sprintf("%s get pod -n %s -l '%s' -o jsonpath='{.items[*].status.phase}'", kubectl, ns, label)
			phase, err := runtime.GetRunner().Host.SudoCmdContext(ctx, cmd, false, true)
			if err != nil {
				return fmt.Errorf("pod status invalid, namespace: %s, label: %s, waiting ...", ns, label)
			}

			if phase != "Running" {
				return fmt.Errorf("pod is %s, namespace: %s, label: %s, waiting ...", phase, ns, label)
			}
		}
	}

	return nil
}

type SetUserInfo struct {
	common.KubeAction
}

func (t *SetUserInfo) Execute(runtime connector.Runtime) error {
	var userName = t.KubeConf.Arg.User.UserName
	var email = t.KubeConf.Arg.User.Email
	var password = t.KubeConf.Arg.User.Password
	var domainName = t.KubeConf.Arg.User.DomainName

	if userName == "" {
		return fmt.Errorf("user info invalid")
	}

	if domainName == "" {
		domainName = cc.DefaultDomainName
	}

	if email == "" {
		email = fmt.Sprintf("%s@%s", userName, domainName)
	}

	if password == "" {
		password, _ = utils.GeneratePassword(8)
	}

	return fmt.Errorf("Not Implemented")
}

type Download struct {
	common.KubeAction
	Version string
	BaseDir string
}

func (t *Download) Execute(runtime connector.Runtime) error {
	if t.KubeConf.Arg.TerminusVersion == "" {
		return errors.New("unknown version to download")
	}

	var fetchMd5 = fmt.Sprintf("curl -sSfL https://dc3p1870nn3cj.cloudfront.net/install-wizard-v%s.md5sum.txt |awk '{print $1}'", t.Version)
	md5sum, err := runtime.GetRunner().Host.Cmd(fetchMd5, false, false)
	if err != nil {
		return errors.New("get md5sum failed")
	}

	var osArch = runtime.GetSystemInfo().GetOsArch()
	var osType = runtime.GetSystemInfo().GetOsType()
	var osVersion = runtime.GetSystemInfo().GetOsVersion()
	var osPlatformFamily = runtime.GetSystemInfo().GetOsPlatformFamily()
	var baseDir = runtime.GetBaseDir()
	var prePath = path.Join(baseDir, "versions")
	var wizard = files.NewKubeBinary("install-wizard", osArch, osType, osVersion, osPlatformFamily, t.Version, prePath)
	wizard.CheckMd5Sum = true
	wizard.Md5sum = md5sum

	if err := wizard.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", wizard.FileName)
	}

	var exists = util.IsExist(wizard.Path())
	if exists {
		if err := wizard.Md5Check(); err == nil {
			// file exists, re-unpack
			return util.Untar(wizard.Path(), wizard.BaseDir)
		} else {
			logger.Error(err)
		}

		util.RemoveFile(wizard.Path())
	}

	logger.Infof("%s downloading %s %s %s ...", common.LocalHost, wizard.ID, wizard.Version)
	if err := wizard.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", wizard.ID, wizard.Url, err)
	}

	return util.Untar(wizard.Path(), wizard.BaseDir)
}

func copyWizard(wizardPath string, np string, runtime connector.Runtime) {
	if util.IsExist(np) {
		util.RemoveDir(np)
	} else {
		// util.Mkdir(np)
	}
	_, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("cp -a %s %s", wizardPath, np), false, false)
	if err != nil {
		logger.Errorf("copy -a %s to %s failed", wizardPath, np)
	}
}

type DownloadFullInstaller struct {
	common.KubeAction
}

func (t *DownloadFullInstaller) Execute(runtime connector.Runtime) error {

	return nil
}

type PrepareFinished struct {
	common.KubeAction
}

func (t *PrepareFinished) Execute(runtime connector.Runtime) error {
	var preparedFile = filepath.Join(runtime.GetBaseDir(), ".prepared")
	return util.WriteFile(preparedFile, []byte(t.KubeConf.Arg.TerminusVersion), cc.FileMode0644)
	// if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("touch %s", preparedFile), false, true); err != nil {
	// 	return err
	// }
	// return nil
}

type CheckPepared struct {
	common.KubeAction
	BaseDir string
	Force   bool
}

func (t *CheckPepared) Execute(runtime connector.Runtime) error {
	var preparedPath = filepath.Join(t.BaseDir, ".prepared")

	if utils.IsExist(preparedPath) {
		t.PipelineCache.Set(common.CachePreparedState, "true") // TODO not used
	} else if t.Force {
		return errors.New("terminus is not prepared well, cannot continue actions")
	}

	return nil
}

type GenerateTerminusUninstallScript struct {
	common.KubeAction
}

func (t *GenerateTerminusUninstallScript) Execute(runtime connector.Runtime) error {
	filePath := path.Join(runtime.GetBaseDir(), uninstalltemplate.TerminusUninstallScriptValues.Name())
	uninstallPath := path.Join("/usr/local/bin", uninstalltemplate.TerminusUninstallScriptValues.Name())
	data := util.Data{
		"BaseDir": runtime.GetBaseDir(),
		"Phase":   "install",
		"Version": t.KubeConf.Arg.TerminusVersion,
	}

	uninstallScriptStr, err := util.Render(uninstalltemplate.TerminusUninstallScriptValues, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render uninstall template failed")
	}

	if err := util.WriteFile(filePath, []byte(uninstallScriptStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write uninstall %s failed", filePath))
	}

	if err := runtime.GetRunner().SudoScp(filePath, uninstallPath); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("scp file %s to remote %s failed", filePath, uninstallPath))
	}

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf %s", filePath), false, false); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("remove file %s failed", filePath))
	}

	return nil
}

type InstallFinished struct {
	common.KubeAction
}

func (t *InstallFinished) Execute(runtime connector.Runtime) error {
	var content = fmt.Sprintf("%s %s", t.KubeConf.Arg.TerminusVersion, t.KubeConf.Arg.Kubetype)
	var phaseState = path.Join(runtime.GetBaseDir(), ".installed")
	if err := util.WriteFile(phaseState, []byte(content), cc.FileMode0644); err != nil {
		return err
	}
	return nil
}

type DeleteWizardFiles struct {
	common.KubeAction
}

func (d *DeleteWizardFiles) Execute(runtime connector.Runtime) error {
	var dirs = []string{
		path.Join(runtime.GetInstallerDir(), "files"),
		path.Join(runtime.GetInstallerDir(), "cli"),
	}

	for _, dir := range dirs {
		if util.IsExist(dir) {
			runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf %s", dir), false, true)
		}
	}
	return nil
}
