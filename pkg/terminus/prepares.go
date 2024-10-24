package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"net"
)

type CheckPublicNetworkInfo struct {
	common.KubePrepare
}

func (p *CheckPublicNetworkInfo) PreCheck(runtime connector.Runtime) (bool, error) {
	if runtime.GetSystemInfo().IsDarwin() {
		return true, nil
	}
	var cmd = "curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-ipv4"
	var publicIp, _ = runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	if publicIp == "" || net.ParseIP(publicIp) == nil {
		cmd = "curl -s --retry 5 --retry-delay 1 --retry-max-time 10 http://ifconfig.me"
		publicIp, _ = runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	}
	if net.ParseIP(publicIp) != nil {
		p.KubeConf.Arg.PublicNetworkInfo.PublicIp = publicIp
	}

	cmd = "curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-hostname"
	var hostname, err = runtime.GetRunner().Host.SudoCmd(cmd, false, false)
	if err == nil {
		p.KubeConf.Arg.PublicNetworkInfo.Hostname = hostname
	}

	return true, nil
}

type NotEqualDesiredVersion struct {
	common.KubePrepare
}

func (n *NotEqualDesiredVersion) PreCheck(runtime connector.Runtime) (bool, error) {
	ksVersion, ok := n.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		ksVersion = ""
	}

	if n.KubeConf.Cluster.KubeSphere.Version == ksVersion {
		return false, nil
	}
	return true, nil
}
