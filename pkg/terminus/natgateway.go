package terminus

import (
	"bytetrade.io/web3os/installer/pkg/utils"
	"fmt"
	"github.com/pkg/errors"
	"net"
	"os"
	"os/exec"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type GetNATGatewayIP struct {
	common.KubeAction
}

func (s *GetNATGatewayIP) Execute(runtime connector.Runtime) error {
	var prompt string
	var input string
	var systemInfo = runtime.GetSystemInfo()
	var hostIP = s.KubeConf.Arg.HostIP

	switch {
	case systemInfo.IsWsl() || systemInfo.IsWindows():
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
		reader, err := utils.GetBufIOReaderOfTerminalInput()
		if err != nil {
			return errors.Wrap(err, "failed to get terminal input reader")
		}
	LOOP:
		fmt.Printf(prompt)
		input, err = reader.ReadString('\n')
		if input == "" {
			if err != nil && err.Error() != "EOF" {
				return err
			}
		}
		input = strings.TrimSpace(input)
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

type UpdateNATGatewayForUser struct {
	common.KubeAction
}

func (a *UpdateNATGatewayForUser) Execute(runtime connector.Runtime) error {
	hostIP := runtime.GetSystemInfo().GetNATGateway()
	if hostIP == "" {
		return errors.New("NAT gateway is not set")
	}
	si := runtime.GetSystemInfo()
	var kubectlCMD string
	var kubectlCMDDefaultArgs []string
	var err error
	if si.IsDarwin() {
		kubectlCMD, err = util.GetCommand(common.CommandKubectl)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "kubectl not found")
		}
	} else if si.IsWindows() {
		kubectlCMD = "cmd"
		kubectlCMDDefaultArgs = []string{"/C", "wsl", "-d", a.KubeConf.Arg.WSLDistribution, "-u", "root", common.CommandKubectl}
	}

	getUserArgs := []string{"get", "user", "-o", "jsonpath='{.items[0].metadata.name}'"}
	getUserCMD := exec.Command(kubectlCMD, append(kubectlCMDDefaultArgs, getUserArgs...)...)
	usernameBytes, err := getUserCMD.Output()
	if err != nil {
		return errors.Wrap(err, "failed to get user for updating")
	}
	username := strings.TrimSpace(string(usernameBytes))
	username = strings.TrimPrefix(username, "'")
	username = strings.TrimSuffix(username, "'")
	if len(username) == 0 {
		return errors.New("failed to get user for updating: got empty username")
	}
	logger.Infof("updating user: %s", username)

	jsonPatch := fmt.Sprintf(`{"metadata":{"annotations":{"bytetrade.io/nat-gateway-ip":"%s"}}}`, hostIP)
	patchUserArgs := []string{"patch", "user", username, "-p", jsonPatch, "--type=merge"}
	patchUserCMD := exec.Command(kubectlCMD, append(kubectlCMDDefaultArgs, patchUserArgs...)...)
	output, err := patchUserCMD.CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "failed to update nat gateway for user, output: %s", output)
	}
	return nil
}
