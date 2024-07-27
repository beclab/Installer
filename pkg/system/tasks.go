package system

import (
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type InstallDeps struct {
	action.BaseAction
}

func (i *InstallDeps) Execute(runtime connector.Runtime) error {
	return nil
}
