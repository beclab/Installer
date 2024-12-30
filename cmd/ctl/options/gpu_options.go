package options

import (
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"github.com/spf13/cobra"
)

type InstallGpuOptions struct {
	Version string
	BaseDir string
}

func NewInstallGpuOptions() *InstallGpuOptions {
	return &InstallGpuOptions{}
}

func (o *InstallGpuOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "The version of the CUDA driver, current supported versions are 12.4")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set Olares package base dir, defaults to $HOME/"+cc.DefaultBaseDir)
}
