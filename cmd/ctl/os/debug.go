package os

import (
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDebugOs() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug Command",
		Run: func(cmd *cobra.Command, args []string) {

			pipelines.DebugCommand()
		},
	}
}
