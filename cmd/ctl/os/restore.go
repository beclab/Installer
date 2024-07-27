package os

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewCmdRestoreOs() *cobra.Command {
	return &cobra.Command{
		Use: "restore",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restore Terminus")
		},
	}
}
