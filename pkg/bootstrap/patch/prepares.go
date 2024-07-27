package patch

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
)

type CheckDepsPrepare struct {
	prepare.BasePrepare
	Command string
}

func (p *CheckDepsPrepare) PreCheck(runtime connector.Runtime) (bool, error) {
	switch constants.OsPlatform {
	case common.Ubuntu, common.Debian, common.Raspbian, common.CentOs, common.Fedora, common.RHEl:
		return false, nil
	}

	if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("command -v %s", p.Command), false, false); err == nil {
		return false, nil
	}

	return true, nil
}
