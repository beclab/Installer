package ctl

import (
	"bytetrade.io/web3os/installer/cmd/ctl/gpu"
	"bytetrade.io/web3os/installer/cmd/ctl/info"
	"bytetrade.io/web3os/installer/cmd/ctl/node"
	"bytetrade.io/web3os/installer/cmd/ctl/os"
	"bytetrade.io/web3os/installer/version"
	"github.com/spf13/cobra"
)

func NewDefaultCommand() *cobra.Command {
	cmds := &cobra.Command{
		Use:               "olares-cli",
		Short:             "Olares Installer",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
		Version:           version.VERSION,
	}

	cmds.AddCommand(info.NewCmdInfo())
	cmds.AddCommand(os.NewCmdOs())
	cmds.AddCommand(node.NewNodeCommand())
	cmds.AddCommand(gpu.NewCmdGpu())

	return cmds
}
