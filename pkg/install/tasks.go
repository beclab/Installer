package install

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type WizardTask struct {
	common.KubeAction
}

// todo
func (t *WizardTask) Execute(runtime connector.Runtime) error {
	return nil
}
