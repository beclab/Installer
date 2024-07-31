package scripts

import (
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

// ~ CopyUninstallScriptModule
// debug
type CopyUninstallScriptModule struct {
	module.BaseTaskModule
}

func (m *CopyUninstallScriptModule) Init() {
	m.Name = "CopyUninstallScript"

	copyUninstallScript := &task.LocalTask{
		Name:   "Copy",
		Action: new(CopyUninstallScriptTask),
	}

	m.Tasks = []task.Interface{
		copyUninstallScript,
	}
}
