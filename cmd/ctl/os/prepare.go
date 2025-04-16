package os

import (
	"log"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

type PrepareSystemOptions struct {
	PrepareOptions *options.CliPrepareSystemOptions
	Components     []string
}

func NewPrepareSystemOptions() *PrepareSystemOptions {
	return &PrepareSystemOptions{
		PrepareOptions: options.NewCliPrepareSystemOptions(),
		Components:     []string{},
	}
}

func NewCmdPrepare() *cobra.Command {
	o := NewPrepareSystemOptions()
	cmd := &cobra.Command{
		Use:   "prepare [component1 component2 ...]",
		Short: "Prepare install",
		Run: func(cmd *cobra.Command, args []string) {
			o.Components = args
			if err := pipelines.PrepareSystemPipeline(o.PrepareOptions, o.Components); err != nil {
				log.Fatalf("error: %v", err)
			}
		},
	}
	o.PrepareOptions.AddFlags(cmd)
	return cmd
}
