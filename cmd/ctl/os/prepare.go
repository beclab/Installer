package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
	"log"
)

type PrepareSystemOptions struct {
	PrepareOptions *options.CliPrepareSystemOptions
}

func NewPrepareSystemOptions() *PrepareSystemOptions {
	return &PrepareSystemOptions{
		PrepareOptions: options.NewCliPrepareSystemOptions(),
	}
}

func NewCmdPrepare() *cobra.Command {
	o := NewPrepareSystemOptions()
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare install",
		Run: func(cmd *cobra.Command, args []string) {

			if err := pipelines.PrepareSystemPipeline(o.PrepareOptions); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.PrepareOptions.AddFlags(cmd)
	return cmd
}
