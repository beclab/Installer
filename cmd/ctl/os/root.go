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
	rootOsCmd.AddCommand(NewCmdPrepare())
	rootOsCmd.AddCommand(NewCmdDownloadWizard())
	rootOsCmd.AddCommand(NewCmdDownload())
	rootOsCmd.AddCommand(NewCmdCheckDownload())
	return rootOsCmd
}
