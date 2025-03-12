package os

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdOs() *cobra.Command {
	rootOsCmd := &cobra.Command{
		Use:   "olares",
		Short: "Operations such as installing and uninstalling Olares can be performed using the --phase parameter, which allows for actions like downloading installation packages, downloading dependencies, and installing Olares.",
	}

	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()

	rootOsCmd.AddCommand(NewCmdPrecheck())
	rootOsCmd.AddCommand(NewCmdRootDownload())
	rootOsCmd.AddCommand(NewCmdPrepare())
	rootOsCmd.AddCommand(NewCmdInstallOs())
	rootOsCmd.AddCommand(NewCmdUninstallOs())
	rootOsCmd.AddCommand(NewCmdChangeIP())
	rootOsCmd.AddCommand(NewCmdRelease())
	rootOsCmd.AddCommand(NewCmdPrintInfo())
	rootOsCmd.AddCommand(NewCmdBackup())

	return rootOsCmd
}
