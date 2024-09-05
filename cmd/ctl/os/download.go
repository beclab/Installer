package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

func NewCmdDownloadWizard() *cobra.Command {
	o := options.NewCliDownloadWizardOptions()
	cmd := &cobra.Command{
		Use:   "download-wizard",
		Short: "Download Terminus Installation Wizard",
		Run: func(cmd *cobra.Command, args []string) {
			helper.InitLog(o.BaseDir)
			if err := pipelines.DownloadInstallationWizard(o); err != nil {
				logger.Fatalf("download terminus installation wizard error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download Terminus Installation Package",
		Run: func(cmd *cobra.Command, args []string) {
			helper.InitLog(o.BaseDir)

			if err := pipelines.DownloadInstallationPackage(o); err != nil {
				logger.Errorf("download terminus installation package error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}

func NewCmdCheckDownload() *cobra.Command {
	o := options.NewCliDownloadOptions()
	cmd := &cobra.Command{
		Use:   "check-download",
		Short: "Check Downloaded Terminus Installation Package",
		Run: func(cmd *cobra.Command, args []string) {
			helper.InitLog(o.BaseDir)

			if err := pipelines.CheckDownloadInstallationPackage(o); err != nil {
				logger.Errorf("check download error: %v", err)
			}
		},
	}

	o.AddFlags(cmd)
	return cmd
}
