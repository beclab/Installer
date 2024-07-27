package install

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"github.com/pkg/errors"
)

// ~ CheckFilesExists
type CheckFilesExists struct {
	common.KubeAction
}

// todo
func (a *CheckFilesExists) Execute(runtime connector.Runtime) error {
	src := runtime.GetWorkDir()
	filePath := fmt.Sprintf("%s/installer/0.1.20/amd64/kk", src)
	if ok := util.IsExist(filePath); !ok {
		// return fmt.Errorf("kk not found")
		return errors.Wrap(errors.New("kk not found"), "")
	}
	return nil
}

// ~ CopyInstallPackage
type CopyInstallPackage struct {
	common.KubeAction
}

func (a *CopyInstallPackage) Execute(runtime connector.Runtime) error {
	src := runtime.GetWorkDir()
	dst := "/tmp/install_log"

	if err := util.RemoveDir(dst); err != nil {
		logger.Warnf("remove dir %s failed %v", dst, err)
	}

	if err := util.Mkdir(dst); err != nil {
		return errors.Wrapf(err, "create path %s failed", dst)
	}

	srcFilePath := fmt.Sprintf("%s/installer/0.1.20/amd64/kk", src)
	if err := util.CopyFile(srcFilePath, fmt.Sprintf("%s/kk", dst)); err != nil {
		return errors.Wrapf(err, "copy file %s to %s failed", srcFilePath, dst)
	}

	fmt.Println("---cip / 1---", src)
	fmt.Println("---cip / 2---", dst)

	return nil
}
