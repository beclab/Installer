package ctl

import (
	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/os"
	"bytetrade.io/web3os/installer/version"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	helper.GetMachineInfo()
	helper.InitLog()

	cmds := &cobra.Command{
		Use:               "Terminus Cli",
		Short:             "Terminus Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
	}

	// cmds.AddCommand(artifact.NewCmdLoadImages())
	// cmds.AddCommand(api.NewCmdApi())
	cmds.AddCommand(os.NewCmdOs())
	// cmds.AddCommand(checksum.NewCmdChecksum())

	return cmds
}
