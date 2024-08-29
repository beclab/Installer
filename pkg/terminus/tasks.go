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
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type CopyToWizard struct {
	common.KubeAction
}

func (t *CopyToWizard) Execute(runtime connector.Runtime) error {
	var terminusComponentsDir = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.ComponentsDir)
	var gpuDir = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.GpuDir)
	var homeComponentsDir = path.Join(runtime.GetRootDir(), cc.ComponentsDir)
	if !util.IsExist(homeComponentsDir) {
		runtime.GetRunner().Host.CmdExt(fmt.Sprintf("cp -a %s %s", terminusComponentsDir, runtime.GetRootDir()), false, true)
	}
	if util.IsExist(gpuDir) && util.CountDirFiles(gpuDir) > 0 {
		runtime.GetRunner().Host.CmdExt(fmt.Sprintf("cp  %s/* %s/", gpuDir, homeComponentsDir), false, true)
	}

	return nil

	// var wizardPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.WizardDir)
	// if !util.IsExist(wizardPath) {
	// 	return nil
	// }

	// var componentsPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir, cc.ComponentsDir)
	// if !util.IsExist(componentsPath) {
	// 	return nil
	// }

	// var homeDir = path.Join("/", "home")
	// homeFiles, err := ioutil.ReadDir(homeDir)
	// if err != nil {
	// 	return nil
	// }

	// var find = false
	// for _, f := range homeFiles {
	// 	if !f.IsDir() {
	// 		continue
	// 	}
	// 	find = true
	// 	var aname = f.Name()
	// 	var np = path.Join("/home", aname, "install-wizard")
	// 	copyWizard(wizardPath, np, runtime)
	// 	var cp = path.Join("/home", aname, "install-wizard", cc.ComponentsDir)
	// 	copyWizard(componentsPath, cp, runtime)

	// 	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("chown -R %s:%s %s", aname, aname, np), false, false); err != nil {
	// 		logger.Errorf("chown %s failed", aname)
	// 	}
	// 	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("chmod +x %s", np), false, false); err != nil {
	// 		logger.Errorf("chmod %s failed", np)
	// 	}
	// }

	// if !find {
	// 	var aname = "home"
	// 	var np = path.Join("/home", aname, "install-wizard")
	// 	copyWizard(wizardPath, np, runtime)
	// 	var cp = path.Join("/home", aname, "install-wizard", cc.ComponentsDir)
	// 	copyWizard(componentsPath, cp, runtime)
	// 	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("chown -R %s:%s %s", aname, aname, np), false, false); err != nil {
	// 		logger.Errorf("chown %s failed", aname)
	// 	}
	// 	if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("chmod +x %s", np), false, false); err != nil {
	// 		logger.Errorf("chmod %s failed", np)
	// 	}
	// }

	// return nil
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
	version string
	BaseDir string
}

func (t *Download) Execute(runtime connector.Runtime) error {
	if t.KubeConf.Arg.TerminusVersion == "" {
		return errors.New("unknown version to download")
	}

	var wizard = files.NewKubeBinary("install-wizard", constants.OsArch, t.version, t.BaseDir)

	if err := wizard.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", wizard.FileName)
	}

	var exists = util.IsExist(wizard.Path())
	if exists {
		util.RemoveFile(wizard.Path())
	}

	if !exists || wizard.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, wizard.ID, wizard.Version)
		if err := wizard.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", wizard.ID, wizard.Url, err)
		}
	}

	util.Untar(wizard.Path(), wizard.BaseDir)

	return nil
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

type TidyInstallerPackage struct {
	common.KubeAction
}

func (t *TidyInstallerPackage) Execute(runtime connector.Runtime) error {
	var preparedLock = path.Join("/var/run/lock/.prepared")
	if util.IsExist(preparedLock) {
		util.RemoveFile(preparedLock)
	}

	var terminusPath = path.Join(runtime.GetHomeDir(), cc.TerminusKey)
	if !util.IsExist(terminusPath) {
		util.Mkdir(terminusPath)
	}

	var currentPkgPath = path.Join(runtime.GetRootDir(), cc.PackageCacheDir)
	if util.CountDirFiles(currentPkgPath) > 0 {
		if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("rm -rf %s/pkg && mv %s %s", terminusPath, currentPkgPath, terminusPath), false, true); err != nil {
			return fmt.Errorf("move pkg %s to %s failed", currentPkgPath, terminusPath)
		}
	}

	var currentComponentsPath = path.Join(runtime.GetRootDir(), cc.ComponentsDir)
	if util.CountDirFiles(currentComponentsPath) > 0 {
		if _, err := runtime.GetRunner().Host.CmdExt(fmt.Sprintf("rm -rf %s/pkg/components && cp -a %s %s/pkg/", terminusPath, currentComponentsPath, terminusPath), false, true); err != nil {
			return fmt.Errorf("copy components %s to %s failed", currentComponentsPath, currentPkgPath)
		}
	}

	var currentImagesPath = path.Join(runtime.GetRootDir(), cc.ImagesDir)
	var cmd = fmt.Sprintf("rm -rf %s/images && mv %s %s && mkdir %s && cp %s/images/images.* %s/", terminusPath, currentImagesPath, terminusPath, currentImagesPath, terminusPath, currentImagesPath)
	if _, err := runtime.GetRunner().Host.CmdExt(cmd, false, true); err != nil {
		return fmt.Errorf("move images %s to %s failed", currentImagesPath, terminusPath)
	}

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
