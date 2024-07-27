package artifact

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"github.com/spf13/cobra"
)

type SetLoadImageOptions struct {
	Options *options.LoadImageOptions
}

func NewLoadImageOptions() *SetLoadImageOptions {
	return &SetLoadImageOptions{
		Options: options.NewLoadImageOptions(),
	}
}

func NewCmdLoadImages() *cobra.Command {
	o := NewLoadImageOptions()
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Load Images",
		Run: func(cmd *cobra.Command, args []string) {
			if err := helper.InitLog(constants.WorkDir); err != nil {
				fmt.Println("init logger failed", err)
				os.Exit(1)
			}

			if err := pipelines.PreloadImages(o.Options.KubeType, o.Options.RegistryMirrors); err != nil {
				logger.Errorf("preload images error %v", err)
			}
		},
	}
	o.Options.AddFlags(cmd)
	return cmd
}
