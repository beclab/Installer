package os

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdOs() *cobra.Command {
	rootOsCmd := &cobra.Command{
		Use:   "terminus",
		Short: "Operations such as installing and uninstalling Terminus can be performed using the --phase parameter, which allows for actions like downloading installation packages, downloading dependencies, and installing Terminus.",
	}

	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()

	// TODO Add a command to install Terminus.
	// TODO Before installing Terminus, we need to obtain user information, the WSL NAT gateway address, etc.

	rootOsCmd.AddCommand(NewCmdRootDownload())
	rootOsCmd.AddCommand(NewCmdPrepare())
	rootOsCmd.AddCommand(NewCmdInstallOs())
	rootOsCmd.AddCommand(NewCmdUninstallOs())
	rootOsCmd.AddCommand(NewCmdPrintInfo())

	return rootOsCmd
}
