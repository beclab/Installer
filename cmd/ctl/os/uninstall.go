package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
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
			logger.Infof("options: minikube: %v, phase: %v, all: %v, storage-type: %s, storage-bucket: %s, base-dir: %s",
				o.UninstallOptions.MiniKube, o.UninstallOptions.Phase, o.UninstallOptions.All,
				o.UninstallOptions.StorageType, o.UninstallOptions.StorageBucket,
				o.UninstallOptions.BaseDir,
			)

			if err := pipelines.UninstallTerminusPipeline(o.UninstallOptions); err != nil {
				logger.Fatalf("delete terminus error %v", err)
			}
		},
	}
	o.UninstallOptions.AddFlags(cmd)
	return cmd
}
