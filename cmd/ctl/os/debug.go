package os

import (
	"fmt"

	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDebugOs() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug Command",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.DebugCommand(); err != nil {
				fmt.Println("---err---", err)
			}
			fmt.Println("---1---")
		},
	}
}
