package os

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/installer/cmd/ctl/helper"
	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/pipelines"
	"bytetrade.io/web3os/installer/version"
	"github.com/spf13/cobra"
)

type InitializeKubeOptions struct {
	InitializeOptions *options.CliKubeInitializeOptions
}

func NewInitializeKubeOptions() *InitializeKubeOptions {
	return &InitializeKubeOptions{
		InitializeOptions: &options.CliKubeInitializeOptions{},
	}
}

func NewCmdInitializeOs() *cobra.Command {
	o := NewInitializeKubeOptions()
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize KubeSphere",
		Run: func(cmd *cobra.Command, args []string) {
			if err := helper.InitLog(constants.WorkDir); err != nil {
				fmt.Println("init logger failed", err)
				os.Exit(1)
			}

			logger.Infof("options: version: %s, kube: %s, minikube: %v, minikubeprofile: %s, registry: %s, k3sctr: %s",
				version.VERSION, o.InitializeOptions.KubeType, o.InitializeOptions.MiniKube,
				o.InitializeOptions.MiniKubeProfile, o.InitializeOptions.RegistryMirrors, o.InitializeOptions.K3sContainerRuntimeEndpoint)

			if err := pipelines.CliInitializeTerminusPipeline(
				o.InitializeOptions.KubeType,
				o.InitializeOptions.MiniKube,
				o.InitializeOptions.MiniKubeProfile,
				o.InitializeOptions.RegistryMirrors,
				o.InitializeOptions.K3sContainerRuntimeEndpoint,
			); err != nil {
				logger.Errorf("initialize kube error: %v", err)
				os.Exit(1)
			}
		},
	}
	o.InitializeOptions.AddFlags(cmd)
	return cmd
}
