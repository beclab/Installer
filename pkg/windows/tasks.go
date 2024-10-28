package windows

import (
	"fmt"
	"path"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

type DownloadImage struct {
	common.KubeAction
}

func (d *DownloadImage) Execute(runtime connector.Runtime) error {
	systemInfo := runtime.GetSystemInfo()
	baseDir := path.Join(runtime.GetSystemInfo().GetHomeDir(), "Documents", "terminus")
	wslImage := files.NewKubeBinary("wsl", systemInfo.GetOsArch(), systemInfo.GetOsType(), systemInfo.GetOsVersion(), systemInfo.GetOsPlatformFamily(),
		d.KubeConf.Arg.TerminusVersion, baseDir)

	if err := wslImage.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", wslImage.FileName)
	}

	if util.IsExist(wslImage.Path()) {
		if err := d.checkWslImageMd5(wslImage.Path(), wslImage.Url); err != nil {
			if err := util.RemoveFile(wslImage.Path()); err != nil {
				return errors.Wrapf(errors.WithStack(err), "remove file %s failed", wslImage.FileName)
			}
		} else {
			return nil
		}
	}

	logger.Debugf("downloading %s ...", wslImage.FileName)
	if err := wslImage.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", wslImage.ID, wslImage.Url, err)
	}

	return nil
}

func (d *DownloadImage) checkWslImageMd5(filePath, urlPath string) error {
	fileMd5, err := utils.FileMD5(filePath)
	if err != nil {
		return err
	}

	var wslImageUrlPath = strings.ReplaceAll(urlPath, ".tar.gz", ".md5sum.txt")
	getMd5, err := utils.NewCommandExecutor([]string{fmt.Sprintf("curl -sSfL %s", wslImageUrlPath)}, false, false).Execute()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("get wsl image md5 failed, url: %s", wslImageUrlPath))
	}
	if !strings.Contains(getMd5, fileMd5) {
		return fmt.Errorf("wsl image md5 not match, fileMd5: %s, remoteMd5: %s", fileMd5, getMd5)
	}
	return nil
}

type ImportImage struct {
	common.KubeAction
}

func (i *ImportImage) Execute(runtime connector.Runtime) error {
	var baseDir = path.Join(runtime.GetSystemInfo().GetHomeDir(), "Documents", "terminus")
	var version = i.KubeConf.Arg.TerminusVersion
	var fileName = path.Join(baseDir, fmt.Sprintf("install-wizard-wsl-image-v%s.tar.gz", version))
	if !util.IsExist(fileName) {
		return fmt.Errorf("%s not found", fileName)
	}

	var distro = fmt.Sprintf("terminus-v%s", version)
	var distroDir = path.Join(baseDir, distro)
	if !util.IsExist(distroDir) {
		util.Mkdir(distroDir)
	}

	var cmd = utils.NewCommandExecutor([]string{"wsl", "--import", distro, distroDir, fileName}, false, false)
	if _, err := cmd.Execute(); err != nil {
		if strings.Contains(err.Error(), "ERROR_ALREADY_EXISTS") {
			return errors.Wrapf(errors.WithStack(err), "DISTRO %s already exists", distro)
		}
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("import image failed"))
	}

	fmt.Printf("Import Distro %s successd\n", distro)
	return nil
}

type ConfigWSLForwardRules struct {
	common.KubeAction
}

func (c *ConfigWSLForwardRules) Execute(runtime connector.Runtime) error {
	var distro = fmt.Sprintf("terminus-v%s", c.KubeConf.Arg.TerminusVersion)
	var cmd = utils.NewCommandExecutor([]string{"wsl", "-d", distro, "bash", "-c", "ip address show eth0 | grep inet | grep -v inet6 | cut -d ' ' -f 6 | cut -d '/' -f 1"}, false, false)

	ip, err := cmd.Execute()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("get wsl %s ip failed", distro))
	}

	logger.Debugf("wsl %s, ip: %s", distro, ip)

	var win = utils.NewCommandExecutor([]string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=80 connectaddress=%s", ip)}, false, false)
	if _, err = win.Execute(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	win = utils.NewCommandExecutor([]string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=%s", ip)}, false, false)
	if _, err = win.Execute(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	win = utils.NewCommandExecutor([]string{fmt.Sprintf("netsh interface portproxy add v4tov4 listenport=30180 listenaddress=0.0.0.0 connectport=30180 connectaddress=%s", ip)}, false, false)
	if _, err = win.Execute(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s forward rules failed", distro))
	}

	return nil
}

type ConfigWSLHostsAndDns struct {
	common.KubeAction
}

func (c *ConfigWSLHostsAndDns) Execute(runtime connector.Runtime) error {
	var distro = fmt.Sprintf("terminus-v%s", c.KubeConf.Arg.TerminusVersion)

	var cmd = utils.NewCommandExecutor([]string{"wsl", "-d", distro, "-u", "root", "bash", "-c", "echo -e '127.0.0.1 localhost\\n$(ip -4 addr show eth0 | grep -oP '(?<=inet\\s)\\d+(\\.\\d+){3}') $(hostname)' > /etc/hosts && echo -e 'nameserver 1.1.1.1\\nnameserver 1.0.0.1' > /etc/resolv.conf"}, false, false)

	if _, err := cmd.Execute(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("config wsl %s hosts and dns failed", distro))
	}

	return nil
}

type InstallTerminus struct {
	common.KubeAction
}

func (i *InstallTerminus) Execute(runtime connector.Runtime) error {
	var distro = fmt.Sprintf("terminus-v%s", i.KubeConf.Arg.TerminusVersion)

	var envs = []string{
		fmt.Sprintf("export %s=%s", common.ENV_KUBE_TYPE, i.KubeConf.Arg.Kubetype),
		fmt.Sprintf("export %s=%s", common.ENV_REGISTRY_MIRRORS, i.KubeConf.Arg.RegistryMirrors),
		fmt.Sprintf("export %s=%d", common.ENV_LOCAL_GPU_ENABLE, utils.FormatBoolToInt(i.KubeConf.Arg.GPU.Enable)),
		fmt.Sprintf("export %s=%d", common.ENV_LOCAL_GPU_SHARE, utils.FormatBoolToInt(i.KubeConf.Arg.GPU.Share)),
		fmt.Sprintf("export %s=%s", common.ENV_CLOUDFLARE_ENABLE, i.KubeConf.Arg.Cloudflare.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_ENABLE, i.KubeConf.Arg.Frp.Enable),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_SERVER, i.KubeConf.Arg.Frp.Server),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_PORT, i.KubeConf.Arg.Frp.Port),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_METHOD, i.KubeConf.Arg.Frp.AuthMethod),
		fmt.Sprintf("export %s=%s", common.ENV_FRP_AUTH_TOKEN, i.KubeConf.Arg.Frp.AuthToken),
		fmt.Sprintf("export %s=%d", common.ENV_TOKEN_MAX_AGE, i.KubeConf.Arg.TokenMaxAge),
		fmt.Sprintf("export %s=%s", common.ENV_MARKET_PROVIDER, i.KubeConf.Arg.MarketProvider),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_CERT_SERVICE_API, i.KubeConf.Arg.TerminusCertServiceAPI),
		fmt.Sprintf("export %s=%s", common.ENV_TERMINUS_DNS_SERVICE_API, i.KubeConf.Arg.TerminusDNSServiceAPI),
	}

	var params = strings.Join(envs, " && ")
	var cmd = utils.NewCommandExecutor([]string{"wsl", "-d", distro, "-u", "ubuntu", "--cd", "/home/ubuntu", "bash", "-c", fmt.Sprintf("%s && bash ~/install.sh", params)}, false, true)
	if _, err := cmd.Execute(); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("install terminus %s failed", distro))
	}
	return nil
}
