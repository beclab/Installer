package terminus

import (
	"fmt"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	uninstalltemplate "bytetrade.io/web3os/installer/pkg/terminus/templates"
	"bytetrade.io/web3os/installer/pkg/utils"

	"github.com/pkg/errors"
)

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
	Md5sum  string
}

func (t *Download) Execute(runtime connector.Runtime) error {
	if t.KubeConf.Arg.TerminusVersion == "" {
		return errors.New("unknown version to download")
	}

	var wizard = files.NewKubeBinary("install-wizard", constants.OsArch, t.Version, t.BaseDir)
	wizard.CheckMd5Sum = true
	wizard.Md5sum = t.Md5sum

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
	BaseDir string
}

func (t *PrepareFinished) Execute(runtime connector.Runtime) error {
	var finPath = filepath.Join(t.BaseDir, ".prepared")
	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("touch %s", finPath), false, true); err != nil {
		return err
	}
	return nil
}

type CheckPepared struct {
	common.KubeAction
	BaseDir string
	Force   bool
}

func (t *CheckPepared) Execute(runtime connector.Runtime) error {
	var finPath = filepath.Join(t.BaseDir, ".prepared")

	if utils.IsExist(finPath) {
		t.PipelineCache.Set(common.CachePreparedState, "true")
	} else if t.Force {
		return errors.New("terminus is not prepared well, cannot continue actions")
	}

	return nil
}

type GenerateTerminusUninstallScript struct {
	common.KubeAction
}

func (t *GenerateTerminusUninstallScript) Execute(runtime connector.Runtime) error {
	uninstallPath := path.Join("/usr/local/bin", uninstalltemplate.TerminusUninstallScriptValues.Name())
	data := util.Data{
		"BaseDir": runtime.GetBaseDir(),
		"Phase":   "install",
	}

	uninstallScriptStr, err := util.Render(uninstalltemplate.TerminusUninstallScriptValues, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render uninstall template failed")
	}

	if err := util.WriteFile(uninstallPath, []byte(uninstallScriptStr), cc.FileMode0755); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write uninstall %s failed", uninstallPath))
	}

	return nil
}

type GenerateInstalledPhaseState struct {
	common.KubeAction
	Phase string
}

func (t *GenerateInstalledPhaseState) Execute(runtime connector.Runtime) error {
	var phaseState = path.Join(runtime.GetBaseDir(), ".installed")
	if err := util.WriteFile(phaseState, nil, cc.FileMode0644); err != nil {
		return err
	}
	return nil
}
