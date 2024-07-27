package dm

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type P1 struct {
	common.KubePrepare
}

func (p *P1) PreCheck(runtime connector.Runtime) (bool, error) {
	// return true, nil
	return false, nil
}

type Task1 struct {
	common.KubeAction
}

func (t *Task1) Execute(runtime connector.Runtime) error {
	fmt.Println("---t1---")
	return nil
}

type Task2 struct {
	common.KubeAction
}

func (t *Task2) Execute(runtime connector.Runtime) error {
	fmt.Println("---t2---")
	return nil
}

type Task3 struct {
	common.KubeAction
}

func (t *Task3) Execute(runtime connector.Runtime) error {
	fmt.Println("---t3---")
	return nil
}
