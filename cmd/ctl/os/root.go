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

	rootOsCmd.AddCommand(NewCmdDownloadWizard())
	rootOsCmd.AddCommand(NewCmdDownload())
	rootOsCmd.AddCommand(NewCmdCheckDownload())
	rootOsCmd.AddCommand(NewCmdPrepare())
	rootOsCmd.AddCommand(NewCmdInstallOs())
	rootOsCmd.AddCommand(NewCmdUninstallOs())
	return rootOsCmd
}
