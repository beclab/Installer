package options

import "github.com/spf13/cobra"

type ChecksumOptions struct {
	File string
}

func NewChecksumOptions() *ChecksumOptions {
	return &ChecksumOptions{}
}

func (o *ChecksumOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.File, "file", "", "file path")
}
