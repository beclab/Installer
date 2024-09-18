package terminus

import (
	"fmt"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type Patch struct {
	common.KubeAction
}

func (t *Patch) Execute(runtime connector.Runtime) error {
	var kubectlpath, _ = util.GetCommand(common.CommandKubectl)
	var installPath = filepath.Dir(t.KubeConf.Arg.Manifest)
	var rolePatchName = "patch-globalrole-workspace-manager.yaml"
	var notificationPatchName = "patch-notification-manager.yaml"
	var rolePatchPath = path.Join(installPath, "deploy", rolePatchName)

	if !util.IsExist(rolePatchPath) {
		return fmt.Errorf("patch %s not exists", rolePatchName)
	}

	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectlpath, rolePatchPath), false, true); err != nil {
		return err
	}

	var notificationPatchPath = path.Join(installPath, "deploy", notificationPatchName)

	if !util.IsExist(notificationPatchPath) {
		return fmt.Errorf("patch %s not exists", notificationPatchName)
	}

	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s apply -f %s", kubectlpath, notificationPatchPath), false, true); err != nil {
		return err
	}

	return nil
}

type TerminusPatchModule struct {
	common.KubeModule
}

func (m *TerminusPatchModule) Init() {

}
