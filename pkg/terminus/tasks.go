package terminus

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type SetUserInfo struct {
	common.KubeAction
}

func (t *SetUserInfo) Execute(runtime connector.Runtime) error {
	var userName = t.KubeConf.Arg.User.UserName
	var email = t.KubeConf.Arg.User.Email
	var password = t.KubeConf.Arg.User.Password
	var domainName = t.KubeConf.Arg.User.DomainName

	if userName == "" {
		return fmt.Errorf("user info invalid")
	}

	if domainName == "" {
		domainName = cc.DefaultDomainName
	}

	if email == "" {
		email = fmt.Sprintf("%s@%s", userName, domainName)
	}

	if password == "" {
		password, _ = utils.GeneratePassword(8)
	}

	return fmt.Errorf("Not Implemented")
}
