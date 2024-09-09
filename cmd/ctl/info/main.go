package info

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/constants"
	"github.com/spf13/cobra"
)

func NewCmdInfo() *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Terminus install, uninstall or restore",
	}
	infoCmd.AddCommand(showInfoCommand())

	return infoCmd
}

func showInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Install Terminus",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`OS_TYPE=%s
OS_PLATFORM=%s
OS_ARCH=%s
OS_VERSION=%s
OS_KERNEL=%s
OS_DETAIL=%s`, constants.OsType, constants.OsPlatform, constants.OsArch, constants.OsVersion, constants.OsKernel, constants.OsDetail)
		},
	}
	return cmd
}
