package terminus

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type GetNATGatewayIP struct {
	common.KubeAction
}

// todo support mac
func (s *GetNATGatewayIP) Execute(runtime connector.Runtime) error {
	var prompt string
	var input string
	var err error
	var systemInfo = runtime.GetSystemInfo()
	var hostIP = s.KubeConf.Arg.HostIP

	switch {
	case systemInfo.IsWsl():
		disableHostIPPrompt := os.Getenv(common.ENV_DISABLE_HOST_IP_PROMPT)
		if strings.EqualFold(disableHostIPPrompt, "") || !util.IsValidIPv4Addr(net.ParseIP(hostIP)) {
			prompt = "Enter the NAT gateway(the Windows host)'s IP [default: " + hostIP + "]: "
		} else {
			input = hostIP
		}
	case systemInfo.IsDarwin():
		prompt = "Enter the NAT gateway(the MacOs host)'s IP: "
	default:
		return nil
	}

	if prompt != "" {
		reader := bufio.NewReader(os.Stdin)
	LOOP:
		fmt.Printf(prompt)
		input, err = reader.ReadString('\n')
		if input == "" {
			if err != nil && err.Error() != "EOF" {
				return err
			}
		}

		input = strings.TrimSuffix(input, "\n")
		if input == "" && hostIP != "" {
			input = hostIP
		}
		if !util.IsValidIPv4Addr(net.ParseIP(input)) {
			fmt.Printf("\nsorry, invalid IP, please try again.\n")
			goto LOOP
		}
	}

	logger.Infof("Nat Gateway IP: %s", input)
	runtime.GetSystemInfo().SetNATGateway(input)
	return nil
}
