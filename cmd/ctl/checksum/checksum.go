package checksum

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/spf13/cobra"
)

type ChecksumOptions struct {
	ChecksumOptions *options.ChecksumOptions
}

func NewChecksumOptions() *ChecksumOptions {
	return &ChecksumOptions{
		ChecksumOptions: options.NewChecksumOptions(),
	}
}

func NewCmdChecksum() *cobra.Command {
	o := NewChecksumOptions()
	cmd := &cobra.Command{
		Use:   "checksum",
		Short: "File checksum verification",
		Run: func(cmd *cobra.Command, args []string) {
			if o.ChecksumOptions.File == "" {
				fmt.Println("file path is required")
				os.Exit(1)
				return
			}

			if ok := utils.IsExist(o.ChecksumOptions.File); !ok {
				fmt.Println("file not found")
				os.Exit(1)
				return
			}
			sum, err := util.Sha256sum(o.ChecksumOptions.File)
			if err != nil {
				fmt.Printf("check sha256sum failed %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("file %s %s\n", o.ChecksumOptions.File, sum)
		},
	}

	o.ChecksumOptions.AddFlags(cmd)

	return cmd
}
