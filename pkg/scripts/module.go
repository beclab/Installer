package scripts

import (
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/task"
)

// ~ CopyUninstallScriptModule
// todo 测试阶段
// ! 测试的，目前在做卸载时，原始的 uninstall_cmd.sh 执行会报错，主要还是执行路径的问题
// ! 这里先拷贝内部嵌入的一个修复版本，等后面拆分脚本时再完善
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
