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
			if err := helper.InitLog(constants.WorkDir); err != nil {
				fmt.Println("init logger failed", err)
				os.Exit(1)
			}

			logger.Infof("options: version: %s, minikube: %v, deletecache: %v, deletecri: %v", o.UninstallOptions.MiniKube, o.UninstallOptions.DeleteCache, o.UninstallOptions.DeleteCRI)

			if err := pipelines.UninstallTerminusPipeline(o.UninstallOptions.MiniKube, o.UninstallOptions.DeleteCache, o.UninstallOptions.DeleteCRI); err != nil {
				logger.Errorf("delete terminus error %v", err)
			}
		},
	}
	o.UninstallOptions.AddFlags(cmd)
	return cmd
}
