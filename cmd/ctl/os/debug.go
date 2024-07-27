package os

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/spf13/cobra"
)

func NewCmdDebugOs() *cobra.Command {
	return &cobra.Command{
		Use:   "debug",
		Short: "Debug Command",
		Run: func(cmd *cobra.Command, args []string) {
			workDir, err := utils.WorkDir()
			if err != nil {
				fmt.Println("working path error", err)
				os.Exit(1)
			}

			constants.WorkDir = workDir

			if err := helper.InitLog(workDir); err != nil {
				fmt.Println("init logger failed", err)
				os.Exit(1)
			}

			if err := pipelines.DebugCommand(); err != nil {
				fmt.Println("---err---", err)
			}
		},
	}
}
