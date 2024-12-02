package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"fmt"
	"time"
)

type WelcomeMessage struct {
	common.KubeAction
}

func (t *WelcomeMessage) Execute(runtime connector.Runtime) error {
	port := 30180
	var ip string
	if runtime.GetSystemInfo().GetLocalIp() != "" {
		ip = runtime.GetSystemInfo().GetLocalIp()
	}
	if si := runtime.GetSystemInfo(); si.GetNATGateway() != "" {
		ip = si.GetNATGateway()
	}
	if ip == "" {
		ip = t.KubeConf.Arg.PublicNetworkInfo.PublicIp
	}

	logger.InfoInstallationProgress("Installation wizard is complete")

	fmt.Printf("\n\n------------------------------------------------\n")
	logger.Info("Olares is running at")
	logger.Infof("http://%s:%d", ip, port)
	logger.Info("Open your browser and visit the above address")
	logger.Infof("Username: %s", t.KubeConf.Arg.User.UserName)
	logger.Infof("Password: %s", t.KubeConf.Arg.User.Password)
	logger.Info("Please change the default password after login")
	fmt.Println("\n------------------------------------------------\n\n")

	logger.InfoInstallationProgress("All done")
	return nil
}

type WelcomeModule struct {
	common.KubeModule
}

func (m *WelcomeModule) Init() {
	logger.InfoInstallationProgress("Starting Olares ...")
	m.Name = "Welcome"

	waitServicesReady := &task.LocalTask{
		Name:   "WaitServicesReady",
		Action: new(CheckKeyPodsRunning),
		Retry:  30,
		Delay:  15 * time.Second,
	}

	welcomeMessage := &task.LocalTask{
		Name:   "WelcomeMessage",
		Action: new(WelcomeMessage),
	}

	m.Tasks = append(m.Tasks, waitServicesReady, welcomeMessage)
}
