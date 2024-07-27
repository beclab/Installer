package ctl

import (
	"bytetrade.io/web3os/installer/cmd/ctl/api"
	"bytetrade.io/web3os/installer/cmd/ctl/artifact"
	"bytetrade.io/web3os/installer/cmd/ctl/checksum"
	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/os"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	helper.GetMachineInfo()

	cmds := &cobra.Command{
		Use:               "Terminus Cli",
		Short:             "Terminus Installer",
		Long:              `......`,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	cmds.AddCommand(artifact.NewCmdLoadImages())
	cmds.AddCommand(api.NewCmdApi())
	cmds.AddCommand(os.NewCmdOs())
	cmds.AddCommand(checksum.NewCmdChecksum())

	return cmds
}
