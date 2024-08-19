package os

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdOs() *cobra.Command {
	rootOsCmd := &cobra.Command{
		Use:   "terminus",
		Short: "Terminus install, uninstall or restore",
	}

	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()

	rootOsCmd.AddCommand(NewCmdInstallOs())
	rootOsCmd.AddCommand(NewCmdInitializeOs())
	rootOsCmd.AddCommand(NewCmdUninstallOs())
	// rootOsCmd.AddCommand(NewCmdChangeIP())
	rootOsCmd.AddCommand(NewCmdDebugOs())
	// rootOsCmd.AddCommand(NewCmdRestoreOs())
	rootOsCmd.AddCommand(NewCmdPrepare())
	return rootOsCmd
}
