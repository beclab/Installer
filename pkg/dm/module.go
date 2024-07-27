package dm

import (
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/kubesphere/plugins"
)

type DmModule struct {
	common.KubeModule
}

func (m *DmModule) Init() {
	m.Name = "dm"
	m.Desc = "dm module"

	checkNodeState := &task.RemoteTask{
		Name:     "CheckNodeState",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(plugins.CheckNodeState),
		Parallel: false,
		Retry:    10,
		Delay:    10 * time.Second,
	}

	// t1 := &task.RemoteTask{
	// 	Name:    "Task1",
	// 	Hosts:   m.Runtime.GetAllHosts(),
	// 	Prepare: &P1{},
	// 	Action:  &Task1{},
	// }

	// t2 := &task.RemoteTask{
	// 	Name:   "Task2",
	// 	Hosts:  m.Runtime.GetAllHosts(),
	// 	Action: &Task2{},
	// }

	// t3 := &task.RemoteTask{
	// 	Name:   "Task3",
	// 	Hosts:  m.Runtime.GetAllHosts(),
	// 	Action: &Task3{},
	// }

	m.Tasks = []task.Interface{
		checkNodeState,
		// t1,
		// t2,
		// t3,
	}
}
