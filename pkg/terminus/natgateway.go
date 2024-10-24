package terminus

import (
	"bufio"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"fmt"
	"net"
	"os"
	"strings"
)

type GetNATGatewayIP struct {
	common.KubeAction
}

func (s *GetNATGatewayIP) Execute(runtime connector.Runtime) error {
	var prompt string
	switch {
	case runtime.GetSystemInfo().IsWsl():
		prompt = "\nEnter the NAT gateway(the Windows host)'s IP: "
	case runtime.GetSystemInfo().IsDarwin():
		prompt = "\nEnter the NAT gateway(the MacOs host)'s IP: "
	default:
		return nil
	}

	reader := bufio.NewReader(os.Stdin)
LOOP:
	fmt.Printf(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	input = strings.TrimSpace(input)
	if input == "" || !util.IsValidIPv4Addr(net.ParseIP(input)) {
		fmt.Printf("\nsorry, invalid IP, please try again")
		goto LOOP
	}

	runtime.GetSystemInfo().SetNATGateway(input)

	return nil
}
