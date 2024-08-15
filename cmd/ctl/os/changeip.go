package os

import (
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

type ChangeIPOptions struct {
	ChangeIPOptions *options.ChangeIPOptions
}

func NewChangeIPOptions() *ChangeIPOptions {
	return &ChangeIPOptions{
		ChangeIPOptions: options.NewChangeIPOptions(),
	}
}

func NewCmdChangeIP() *cobra.Command {
	o := NewChangeIPOptions()
	cmd := &cobra.Command{
		Use:   "changeip",
		Short: "Change IP",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Infof("options: last-ip: %s", o.ChangeIPOptions.LastIP)

			if err := pipelines.ChangeClusterIPPipeline(o.ChangeIPOptions); err != nil {
				logger.Errorf("change cluster ip error: %v", err)
			}
		},
	}

	o.ChangeIPOptions.AddFlags(cmd)
	return cmd
}
