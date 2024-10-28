package cluster

import (
	"bytetrade.io/web3os/installer/pkg/core/module"
)

type phase []module.Module

func (p phase) addModule(m ...module.Module) phase {
	return append(p, m...)
}
