package scripts

import (
	"fmt"

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
