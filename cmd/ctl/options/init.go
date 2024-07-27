package options

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/constants"
)

func InitEnv(o *ApiOptions) {
	fmt.Println(constants.Logo)

	constants.ApiServerListenAddress = o.Port
	constants.Proxy = o.Proxy
}
