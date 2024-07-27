package os

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/spf13/cobra"
)

type UninstallOsOptions struct {
	UninstallOptions *options.CliTerminusUninstallOptions
}

func NewUninstallOsOptions() *UninstallOsOptions {
	return &UninstallOsOptions{
		UninstallOptions: options.NewCliTerminusUninstallOptions(),
	}
}

func NewCmdUninstallOs() *cobra.Command {
	o := NewUninstallOsOptions()
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall Terminus",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(constants.Logo)

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

			if err := pipelines.UninstallTerminusPipeline(); err != nil {
				logger.Errorf("delete terminus error %v", err)
			}
		},
	}
	o.UninstallOptions.AddFlags(cmd)
	return cmd
}
