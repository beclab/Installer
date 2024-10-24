package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"fmt"
	"time"
)

type WelcomeServicePrecheck struct {
	common.KubeAction
}

func (t *WelcomeServicePrecheck) Execute(runtime connector.Runtime) error {
	action := CheckPodsRunning{
		labels: map[string][]string{
			fmt.Sprintf("user-space-%s", t.KubeConf.Arg.User.UserName): {
				"app=edge-desktop",
				"app=vault",
				"app=wizard",
				"app=settings",
				"app=system-frontend",
				"app=authelia",
				"tier=bfl",
			},
		},
	}
	return action.Execute(runtime)
}

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
	fmt.Println("\n\n------------------------------------------------\n")
	fmt.Println("Installation is completed")
	fmt.Println("Terminus is running at")
	fmt.Printf("http://%s:%d\n", ip, port)
	fmt.Println("Open your browser and visit the above address")
	fmt.Printf("Username: %s\n", t.KubeConf.Arg.User.UserName)
	fmt.Printf("Password: %s\n", t.KubeConf.Arg.User.Password)
	fmt.Println("Please change the default password after login")
	fmt.Println("\n------------------------------------------------\n\n")
	return nil
}

type WelcomeModule struct {
	common.KubeModule
}

func (m *WelcomeModule) Init() {
	m.Name = "Welcome"

	waitServicesReady := &task.LocalTask{
		Name:   "WaitServicesReady",
		Action: new(WelcomeServicePrecheck),
		Retry:  30,
		Delay:  10 * time.Second,
	}

	welcomeMessage := &task.LocalTask{
		Name:   "WelcomeMessage",
		Action: new(WelcomeMessage),
	}

	m.Tasks = append(m.Tasks, waitServicesReady, welcomeMessage)
}
