package api

import (
	"os"
	"os/user"

	"bytetrade.io/web3os/installer/cmd/ctl/options"
	"bytetrade.io/web3os/installer/pkg/apiserver"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/phase/startup"
	"github.com/spf13/cobra"
)

type ApiServerOptions struct {
	ApiOptions *options.ApiOptions
}

func NewApiServerOptions() *ApiServerOptions {
	return &ApiServerOptions{
		ApiOptions: options.NewApiOptions(),
	}
}

func NewCmdApi() *cobra.Command {
	o := NewApiServerOptions()
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Terminus Api Server",
		PreRun: func(cmd *cobra.Command, args []string) {
			options.InitEnv(o.ApiOptions)
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := GetCurrentUser(); err != nil {
				logger.Errorf(err.Error())
				os.Exit(1)
			}

			logger.Debugf("current user: %s", constants.CurrentUser)

			if constants.CurrentUser != "root" {
				logger.Error("Current user is not root!! exit ...")
				os.Exit(0)
			}

			logger.Info("Terminus Installer starting ...")

			if err := GetMachineInfo(); err != nil {
				logger.Errorf("failed to get machine info: %+v", err)
				os.Exit(1)
			}

			if err := Run(o.ApiOptions); err != nil {
				logger.Errorf("failed to run installer api server: %+v", err)
				os.Exit(1)
			}
		},
	}

	o.ApiOptions.AddFlags(cmd)

	return cmd
}

func GetCurrentUser() error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	constants.CurrentUser = u.Username
	return nil
}

func GetMachineInfo() error {
	if err := startup.GetMachineInfo(); err != nil {
		return err
	}

	return nil
}

func Run(option *options.ApiOptions) error {

	logger.Infow("[Installer] API Server startup flags",
		"enabled", option.Enabled,
		"port", option.Port,
	)

	s, err := apiserver.New()
	if err != nil {
		return err
	}

	if err = s.PrepareRun(); err != nil {
		return err
	}

	return s.Run()
}
