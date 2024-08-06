package terminus

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
)

type CheckAwsHost struct {
	common.KubePrepare
}

func (p *CheckAwsHost) PreCheck(runtime connector.Runtime) (bool, error) {
	var cmd = "curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-ipv4"
	var publicIp, _ = runtime.GetRunner().SudoCmdExt(cmd, false, false)

	cmd = "curl --connect-timeout 5 -sL http://169.254.169.254/latest/meta-data/public-hostname"
	var hostname, _ = runtime.GetRunner().SudoCmdExt(cmd, false, false)

	p.KubeConf.Arg.AWS.PublicIp = publicIp
	p.KubeConf.Arg.AWS.Hostname = hostname

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
