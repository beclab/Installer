package ctl

import (
	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/info"
	"bytetrade.io/web3os/installer/cmd/ctl/os"
	"bytetrade.io/web3os/installer/version"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	helper.GetMachineInfo()

	cmds := &cobra.Command{
		Use:               "Terminus Cli",
		Short:             "Terminus Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
	}

	cmds.AddCommand(info.NewCmdInfo())
	cmds.AddCommand(os.NewCmdOs())

	return cmds
}
