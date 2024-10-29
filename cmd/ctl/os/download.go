package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdRootDownload() *cobra.Command {
	rootDownloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download the packages and components needed to install Terminus",
	}

	rootDownloadCmd.AddCommand(NewCmdCheckDownload())
	rootDownloadCmd.AddCommand(NewCmdDownload())
	rootDownloadCmd.AddCommand(NewCmdDownloadWizard())

	return rootDownloadCmd
}

func NewCmdDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "component",
		Short: "Download the packages and components needed to install Terminus",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationPackage(o); err != nil {
				logger.Fatalf("download terminus installation package error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdDownloadWizard() *cobra.Command {
	o := options.NewCliDownloadWizardOptions()
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Download the Terminus installation wizard",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.DownloadInstallationWizard(o); err != nil {
				logger.Fatalf("download terminus installation wizard error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdCheckDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check Downloaded Terminus Installation Package",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.CheckDownloadInstallationPackage(o); err != nil {
				logger.Errorf("check download error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}
