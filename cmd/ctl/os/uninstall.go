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
			logger.Infof("options: version: %s, minikube: %v, deletecache: %v, deletecri: %v, storage-type: %s, storage-bucket: %s",
				o.UninstallOptions.MiniKube, o.UninstallOptions.DeleteCache, o.UninstallOptions.DeleteCRI, o.UninstallOptions.StorageType, o.UninstallOptions.StorageBucket)

			if err := pipelines.UninstallTerminusPipeline(o.UninstallOptions); err != nil {
				logger.Errorf("delete terminus error %v", err)
			}
		},
	}
	o.UninstallOptions.AddFlags(cmd)
	return cmd
}
