package windows

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	templates "bytetrade.io/web3os/installer/pkg/windows/templates"
	"github.com/pkg/errors"
)

const (
	windowsAppPath = "AppData\\Local\\Microsoft\\WindowsApps"
	// ubuntu22exe    = "ubuntu2204.exe"
	ubuntuexe = "ubuntu.exe"

	OLARES_WINDOWS_FIREWALL_RULE_NAME = "OlaresRule"
)

var ubuntuTool string
var distro string

type AddAppxPackage struct {
	common.KubeAction
}

func (i *AddAppxPackage) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	// var windowsAppsPath = fmt.Sprintf("%s\\%s", runtime.GetSystemInfo().GetHomeDir(), windowsAppPath)

	// if utils.IsExist(fmt.Sprintf("%s\\%s", windowsAppsPath, ubuntu22exe)) {
	// 	ubuntuTool = ubuntu22exe
	// 	distro = "Ubuntu-22.04"
	// 	return nil
	// }

	appx := files.NewKubeBinary("wsl", systemInfo.GetOsArch(), systemInfo.GetOsType(), systemInfo.GetOsVersion(), systemInfo.GetOsPlatformFamily(), "2204", fmt.Sprintf("%s\\%s\\%s\\%s", systemInfo.GetHomeDir(), cc.DefaultBaseDir, "pkg", "components"), cc.DownloadUrl)

	if err := appx.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", appx.FileName)
	}

	var exists = util.IsExist(appx.Path())
	if exists {
		p := appx.Path()
		output := util.LocalMd5Sum(p)
		if output != appx.Md5sum {
			util.RemoveFile(p)
		}
	}

	if !exists {
		if err := appx.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", appx.ID, appx.Url, err)
		}
	}

	var ps = &utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Add-AppxPackage %s -ForceUpdateFromAnyVersion", appx.Path())},
	}

	if _, err := ps.Run(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Add appx package %s failed", appx.Path()))
	}

	ubuntuTool = ubuntuexe
	distro = "Ubuntu"

	return nil
}

type UpdateWSL struct {
	common.KubeAction
}

func (u *UpdateWSL) Execute(runtime connector.Runtime) error {
	var wslConfigFile = fmt.Sprintf("%s\\%s", runtime.GetSystemInfo().GetHomeDir(), templates.WSLConfigValue.Name())

	file, err := os.Create(wslConfigFile)
	defer file.Close()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("create wsl config %s failed", wslConfigFile))
	}

	systemInfo := runtime.GetSystemInfo()
	memory := u.getMemroy(systemInfo.GetTotalMemory())
	var data = util.Data{
		"Memory": memory,
	}

	wslConfigStr, err := util.Render(templates.WSLConfigValue, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render account template failed")
	}

	if _, err = file.WriteString(wslConfigStr); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write wsl config %s failed", wslConfigFile))
	}

	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "--update"},
	}
	if _, err := cmd.Run(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("Update WSL failed"))
	}

	return nil
}

func (u *UpdateWSL) getMemroy(total uint64) uint64 {
	var memory uint64 = 12
	m := os.Getenv("WSL_MEMORY")
	if m == "" {
		return memory
	}

	sets, err := strconv.ParseUint(m, 10, 64)
	if err != nil {
		return memory
	}

	localMemeory := total / 1024 / 1024 / 1024
	if localMemeory < sets {
		if localMemeory > memory {
			return memory
		} else {
			return localMemeory
		}
	}

	return sets
}

type InstallWSLDistro struct {
	common.KubeAction
}

func (i *InstallWSLDistro) Execute(runtime connector.Runtime) error {
	var cmd = &utils.PowerShellCommandExecutor{
		Commands:  []string{ubuntuTool, "install", "--root"},
		PrintLine: true,
	}
	if _, err := cmd.Run(); err != nil {
		fmt.Printf("Install Ubuntu failed, please check if %s is already installed.\nyou can uninstall it by \"wsl --unregister <Distro>\".\n\n", distro)
		return err
	}

	logger.Infof("Install Ubuntu Distro %s successd\n", distro)

	return nil
}

type ConfigWslConf struct {
	common.KubeAction
}

func (c *ConfigWslConf) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "-d", distro, "-u", "root", "bash", "-c", "echo -e '[boot]\\nsystemd=true\\ncommand=\"mount --make-rshared /\"\\n[network]\\ngenerateHosts=false\\ngenerateResolvConf=false\\nhostname=terminus' > /etc/wsl.conf"},
	}
	if _, err := cmd.Run(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s hosts and dns failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "--shutdown", distro},
	}
	if _, err := cmd.Run(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("shutdown wsl %s failed", distro))
	}

	return nil
}

type ConfigWSLForwardRules struct {
	common.KubeAction
}

func (c *ConfigWSLForwardRules) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "-d", distro, "bash", "-c", "ip address show eth0 | grep inet | grep -v inet6 | cut -d ' ' -f 6 | cut -d '/' -f 1"},
	}

	ip, err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("get wsl %s ip failed", distro))
	}

	logger.Infof("wsl %s, ip: %s", distro, ip)

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=80 connectaddress=%s", ip)},
	}

	if _, err = cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 80 failed, maybe it's already exist %v", err)
		// return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=%s", ip)},
	}
	if _, err = cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 443 failed, maybe it's already exist %v", err)
		// return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=30180 listenaddress=0.0.0.0 connectport=30180 connectaddress=%s", ip)},
	}

	if _, err = cmd.Run(); err != nil {
		logger.Debugf("set portproxy listenport 30180 failed, maybe it's already exist %v", err)
		// return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	return nil
}

type ConfigWSLHostsAndDns struct {
	common.KubeAction
}

func (c *ConfigWSLHostsAndDns) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "-d", distro, "-u", "root", "bash", "-c", "chattr -i /etc/hosts /etc/resolv.conf && "},
	}
	_, _ = cmd.Run()

	cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "-d", distro, "-u", "root", "bash", "-c", "echo -e '127.0.0.1 localhost\\n$(ip -4 addr show eth0 | grep -oP '(?<=inet\\s)\\d+(\\.\\d+){3}') $(hostname)' > /etc/hosts && echo -e 'nameserver 1.1.1.1\\nnameserver 1.0.0.1' > /etc/resolv.conf"},
	}

	if _, err := cmd.Run(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s hosts and dns failed", distro))
	}

	return nil
}

type ConfigWindowsFirewallRule struct {
	common.KubeAction
}

func (c *ConfigWindowsFirewallRule) Execute(runtime connector.Runtime) error {
	var setFirewallRule bool = false
	var autoSetFirewallRule = os.Getenv(common.ENV_AUTO_ADD_FIREWALL_RULES)
	autoSetFirewallRule = strings.TrimSpace(autoSetFirewallRule)

	switch {
	case autoSetFirewallRule != "":
		setFirewallRule = true
		break
	default:
		scanner := bufio.NewScanner(os.Stdin)

		for {
			fmt.Print("\nAccessing Olares requires setting up firewall rules, specifically adding TCP inbound rules for ports 80, 443, and 30180.\nDo you want to set up the firewall rules? (yes/no): ")
			scanner.Scan()
			confirmation := scanner.Text()
			confirmation = strings.TrimSpace(confirmation)
			confirmation = strings.ToLower(confirmation)

			switch confirmation {
			case "y", "yes":
				setFirewallRule = true
				break
			case "n", "no":
				break
			default:
				continue
			}
			break
		}
	}

	if !setFirewallRule {
		fmt.Printf("\nFirewall settings have been skipped. \nIf you want to access the Olares application, please go to the Windows Defender Firewall rules and add an inbound rule for TCP protocol with port numbers 80, 443, and 30180.\n\n\n")
		return nil
	}

	var ps = &utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Get-NetFirewallRule | Where-Object { $_.DisplayName -eq \"%s\" -and $_.Enabled -eq 'True'} | Get-NetFirewallPortFilter | Where-Object { $_.LocalPort -eq 80 -and $_.LocalPort -eq 443 -and $_.LocalPort -eq 30180 -and $_.Protocol -eq 'TCP' } ", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
	}
	rules, _ := ps.Run()
	rules = strings.TrimSpace(rules)
	if rules == "" {
		ps = &utils.PowerShellCommandExecutor{
			Commands: []string{fmt.Sprintf("New-NetFirewallRule -DisplayName \"%s\" -Direction Inbound -Protocol TCP -LocalPort 80,443,30180 -Action Allow", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
		}
		if _, err := ps.Run(); err != nil {
			return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config windows firewall rule %s failed", OLARES_WINDOWS_FIREWALL_RULE_NAME))
		}
	}

	return nil
}

type InstallTerminus struct {
	common.KubeAction
}

func (i *InstallTerminus) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	// var windowsUserPath = convertPath(systemInfo.GetHomeDir())

	var envs = []string{
		fmt.Sprintf("export %s=%s", common.ENV_KUBE_TYPE, i.KubeConf.Arg.Kubetype),
		fmt.Sprintf("export %s=%s", common.ENV_REGISTRY_MIRRORS, i.KubeConf.Arg.RegistryMirrors),
		fmt.Sprintf("export %s=%s", common.ENV_CLOUDFLARE_ENABLE, i.KubeConf.Arg.Cloudflare.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_ENABLE, i.KubeConf.Arg.Frp.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_SERVER, i.KubeConf.Arg.Frp.Server),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_PORT, i.KubeConf.Arg.Frp.Port),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_METHOD, i.KubeConf.Arg.Frp.AuthMethod),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_TOKEN, i.KubeConf.Arg.Frp.AuthToken),
		fmt.Sprintf("export %s=%d", common.ENV_TOKEN_MAX_AGE, i.KubeConf.Arg.TokenMaxAge),
		fmt.Sprintf("export %s=%s", common.ENV_PREINSTALL, os.Getenv(common.ENV_PREINSTALL)),
		fmt.Sprintf("export %s=%s", common.ENV_MARKET_PROVIDER, i.KubeConf.Arg.MarketProvider),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_CERT_SERVICE_API, i.KubeConf.Arg.TerminusCertServiceAPI),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_DNS_SERVICE_API, i.KubeConf.Arg.TerminusDNSServiceAPI),
		fmt.Sprintf("export %s=%s", common.ENV_HOST_IP, systemInfo.GetLocalIp()),
		fmt.Sprintf("export %s=%s", common.ENV_DISABLE_HOST_IP_PROMPT, os.Getenv(common.ENV_DISABLE_HOST_IP_PROMPT)),
		fmt.Sprintf("export %s=%s", common.ENV_DOWNLOAD_CDN_URL, i.KubeConf.Arg.DownloadCdnUrl),
	}

	for key, val := range common.TerminusGlobalEnvs {
		envs = append(envs, fmt.Sprintf("export %s=%s", key, val))
	}

	var installScript = fmt.Sprintf("curl -fsSL https://olares.sh | bash -")
	if i.KubeConf.Arg.TerminusVersion != "" {
		var installFile = fmt.Sprintf("install-wizard-v%s.tar.gz", i.KubeConf.Arg.TerminusVersion)
		installScript = fmt.Sprintf("curl -fsSLO %s/%s && tar -xf %s -C ./ ./install.sh && rm -rf %s && bash ./install.sh",
			cc.DownloadUrl, installFile, installFile, installFile)
	}

	var params = strings.Join(envs, " && ")

	var cmd = &utils.DefaultCommandExecutor{
		Commands:  []string{"wsl", "-d", distro, "-u", "root", "--cd", "/root", "bash", "-c", fmt.Sprintf("%s && %s", params, installScript)},
		PrintLine: true,
	}

	if _, err := cmd.Exec(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install Olares %s failed", distro))
	}

	exec.Command("cmd", "/C", "wsl", "-d", distro, "--exec", "dbus-launch", "true").Run()

	return nil
}

func convertPath(windowsPath string) string {
	linuxPath := strings.ReplaceAll(windowsPath, `\`, `/`)
	if len(linuxPath) > 1 && linuxPath[1] == ':' {
		drive := strings.ToLower(string(linuxPath[0]))
		linuxPath = "/mnt/" + drive + linuxPath[2:]
	}

	return linuxPath
}

type UninstallOlares struct {
	common.KubeAction
}

func (u *UninstallOlares) Execute(runtime connector.Runtime) error {
	var cmd = &utils.DefaultCommandExecutor{
		Commands: []string{"wsl", "--unregister", "Ubuntu"},
	}
	_, _ = cmd.Run()

	return nil
}

type RemoveFirewallRule struct {
	common.KubeAction
}

func (r *RemoveFirewallRule) Execute(runtime connector.Runtime) error {
	(&utils.PowerShellCommandExecutor{
		Commands: []string{fmt.Sprintf("Remove-NetFirewallRule -DisplayName \"%s\"", OLARES_WINDOWS_FIREWALL_RULE_NAME)},
	}).Run()

	return nil
}

type RemovePortProxy struct {
	common.KubeAction
}

func (r *RemovePortProxy) Execute(runtime connector.Runtime) error {
	var ports = []string{"80", "443", "30180"}
	for _, port := range ports {
		(&utils.DefaultCommandExecutor{
			Commands: []string{fmt.Sprintf("netsh interface portproxy delete v4tov4 listenport=%s listenaddress=0.0.0.0", port)}}).Run()
	}

	return nil
}
