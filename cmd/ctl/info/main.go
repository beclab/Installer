package info

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/constants"
	"github.com/spf13/cobra"
)

func NewCmdInfo() *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Print system information, etc.",
	}
	infoCmd.AddCommand(showInfoCommand())

	return infoCmd
}

func showInfoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print system information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf(`OS_TYPE=%s
OS_PLATFORM=%s
OS_ARCH=%s
OS_VERSION=%s
OS_KERNEL=%s
OS_INFO=%s
`, constants.OsType, constants.OsPlatform, constants.OsArch, constants.OsVersion, constants.OsKernel, constants.OsInfo)
		},
	}
	return cmd
}
