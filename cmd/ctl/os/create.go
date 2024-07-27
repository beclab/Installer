package os

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

type InstallOsOptions struct {
	InstallOptions *options.CliTerminusInstallOptions
}

func NewInstallOsOptions() *InstallOsOptions {
	return &InstallOsOptions{
		InstallOptions: options.NewCliTerminusInstallOptions(),
	}
}

func NewCmdInstallOs() *cobra.Command {
	o := NewInstallOsOptions()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Install Terminus",
		Run: func(cmd *cobra.Command, args []string) {
			if err := helper.InitLog(constants.WorkDir); err != nil {
				fmt.Println("init logger failed", err)
				os.Exit(1)
			}

			if err := pipelines.CliInstallTerminusPipeline(o.InstallOptions.KubeType, o.InstallOptions.Proxy); err != nil {
				logger.Errorf("install terminus error %v", err)
			}
		},
	}
	o.InstallOptions.AddFlags(cmd)
	return cmd
}
