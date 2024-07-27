package scripts

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

// ~ Greeting
type Greeting struct {
	action.BaseAction
}

func (t *Greeting) Execute(runtime connector.Runtime) error {
	p := fmt.Sprintf("%s/%s/%s", constants.WorkDir, common.ScriptsDir, common.GreetingShell)
	if ok := util.IsExist(p); ok {
		outstd, _, err := util.Exec(p, false, false)
		if err != nil {
			return err
		}
		logger.Debugf("script CMD: %s, OUTPUT: \n%s", p, outstd)
	}
	return nil
}

// ~ CopyUninstallScriptTask
type CopyUninstallScriptTask struct {
	action.BaseAction
}

func (t *CopyUninstallScriptTask) Execute(runtime connector.Runtime) error {
	dest := path.Join(runtime.GetPackageDir(), common.InstallDir)

	if ok := util.IsExist(dest); !ok {
		return fmt.Errorf("directory %s not exist", dest)
	}

	all := Assets()
	fileContent, err := all.ReadFile(path.Join("files", common.UninstallOsScript))
	if err != nil {
		return fmt.Errorf("read file %s failed: %v", common.UninstallOsScript, err)
	}

	dstFile := path.Join(dest, common.UninstallOsScript)
	err = ioutil.WriteFile(dstFile, fileContent, common.FileMode0755)
	if err != nil {
		log.Fatalf("failed to write file %s to target directory: %v", dstFile, err)
	}

	return nil
}
