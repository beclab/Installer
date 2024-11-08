package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
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
		Use:   "install",
		Short: "Install Terminus",
		Run: func(cmd *cobra.Command, args []string) {
			if err := pipelines.CliInstallTerminusPipeline(o.InstallOptions); err != nil {
				logger.Fatalf("install Olares error: %v", err)
			}
		},
	}
	o.InstallOptions.AddFlags(cmd)
	return cmd
}
