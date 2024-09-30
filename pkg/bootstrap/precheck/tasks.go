/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package precheck

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/action"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"bytetrade.io/web3os/installer/pkg/version/kubernetes"
	"bytetrade.io/web3os/installer/pkg/version/kubesphere"
	"github.com/pkg/errors"
	versionutil "k8s.io/apimachinery/pkg/util/version"
)

type CorrectHostname struct {
	common.KubeAction
}

func (t *CorrectHostname) Execute(runtime connector.Runtime) error {
	hostName := runtime.GetSystemInfo().GetHostname()
	if !utils.ContainsUppercase(hostName) {
		return nil
	}
	hostname := strings.ToLower(hostName)
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("hostnamectl set-hostname %s", hostname), false, true); err != nil {
		return err
	}
	runtime.GetSystemInfo().SetHostname(hostname)
	return nil
}

type RaspbianCheckTask struct {
	common.KubeAction
}

func (t *RaspbianCheckTask) Execute(runtime connector.Runtime) error {
	// if util.IsExist(common.RaspbianCmdlineFile) || util.IsExist(common.RaspbianFirmwareFile) {
	systemInfo := runtime.GetSystemInfo()
	if systemInfo.IsRaspbian() {
		if _, err := util.GetCommand(common.CommandIptables); err != nil {
			_, err = runtime.GetRunner().Host.SudoCmd("apt install -y iptables", false, false)
			if err != nil {
				logger.Errorf("%s install iptables error %v", common.Raspbian, err)
				return err
			}

			_, err = runtime.GetRunner().Host.Cmd("systemctl disable --user gvfs-udisks2-volume-monitor", false, true)
			if err != nil {
				logger.Errorf("%s exec error %v", common.Raspbian, err)
				return err
			}

			_, err = runtime.GetRunner().Host.Cmd("systemctl stop --user gvfs-udisks2-volume-monitor", false, true)
			if err != nil {
				logger.Errorf("%s exec error %v", common.Raspbian, err)
				return err
			}

			if !systemInfo.CgroupCpuEnabled() || !systemInfo.CgroupMemoryEnabled() {
				return fmt.Errorf("cpu or memory cgroups disabled, please edit /boot/cmdline.txt or /boot/firmware/cmdline.txt and reboot to enable it")
			}
		}
	}
	return nil
}

type DisableLocalDNSTask struct {
	common.KubeAction
}

func (t *DisableLocalDNSTask) Execute(runtime connector.Runtime) error {
	switch runtime.GetSystemInfo().GetOsPlatformFamily() {
	case common.Ubuntu, common.Debian:
		stdout, _ := runtime.GetRunner().Host.SudoCmd("systemctl is-active systemd-resolved", false, false)
		if stdout != "active" {
			_, _ = runtime.GetRunner().Host.SudoCmd("systemctl stop systemd-resolved.service", false, true)
			_, _ = runtime.GetRunner().Host.SudoCmd("systemctl disable systemd-resolved.service", false, true)

			if utils.IsExist("/usr/bin/systemd-resolve") {
				_, _ = runtime.GetRunner().Host.SudoCmd("mv /usr/bin/systemd-resolve /usr/bin/systemd-resolve.bak", false, true)
			}
			ok, err := utils.IsSymLink("/etc/resolv.conf")
			if err != nil {
				logger.Errorf("check /etc/resolv.conf error %v", err)
				return err
			}
			if ok {
				if _, err := runtime.GetRunner().Host.SudoCmd("unlink /etc/resolv.conf && touch /etc/resolv.conf", false, true); err != nil {
					logger.Errorf("unlink /etc/resolv.conf error %v", err)
					return err
				}
			}

			if err = ConfigResolvConf(runtime); err != nil {
				logger.Errorf("config /etc/resolv.conf error %v", err)
				return err
			}
		} else {
			if _, err := runtime.GetRunner().Host.SudoCmd("cat /etc/resolv.conf > /etc/resolv.conf.bak", false, true); err != nil {
				logger.Errorf("backup /etc/resolv.conf error %v", err)
				return err
			}
		}
	}

	sysInfo := runtime.GetSystemInfo()
	localIp := sysInfo.GetLocalIp()
	hostname := sysInfo.GetHostname()
	if stdout, _ := runtime.GetRunner().Host.SudoCmd("hostname -i &>/dev/null", false, true); stdout == "" {
		if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("echo %s %s >> /etc/hosts", localIp, hostname), false, true); err != nil {
			return err
		}
	}

	httpCode, _ := utils.GetHttpStatus("https://download.docker.com/linux/ubuntu")
	if httpCode != 200 {
		if err := ConfigResolvConf(runtime); err != nil {
			logger.Errorf("config /etc/resolv.conf error %v", err)
			return err
		}
		if utils.IsExist("/etc/resolv.conf.bak") {
			if err := utils.DeleteFile("/etc/resolv.conf.bak"); err != nil {
				logger.Errorf("remove /etc/resolv.conf.bak error %v", err)
				return err
			}
		}
	}

	return nil
}

func ConfigResolvConf(runtime connector.Runtime) error {
	var err error
	var cmd string

	if constants.CloudVendor == common.AliYun {
		cmd = `echo "nameserver 100.100.2.136" > /etc/resolv.conf`
		if _, err = runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
			logger.Errorf("exec %s error %v", cmd, err)
			return err
		}
	}

	cmd = `echo "nameserver 1.0.0.1" >> /etc/resolv.conf`
	if _, err = runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("exec %s error %v", cmd, err)
		return err
	}

	cmd = `echo "nameserver 1.1.1.1" >> /etc/resolv.conf`
	if _, err = runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
		logger.Errorf("exec %s error %v", cmd, err)
		return err
	}
	return nil
}

type GetSysInfoTask struct {
	action.BaseAction
}

func (t *GetSysInfoTask) Execute(runtime connector.Runtime) error {
	// logger.Infof("os info, all: %s", constants.OsInfo)
	// logger.Infof("host info, user: %s, hostname: %s, hostid: %s, os: %s, platform: %s, version: %s, arch: %s",
	// 	constants.CurrentUser, constants.HostName, constants.HostId, constants.OsType, constants.OsPlatform, constants.OsVersion, constants.OsArch)
	// logger.Infof("kernel info, version: %s", constants.OsKernel)
	// logger.Infof("virtual info, role: %s, system: %s", constants.VirtualizationRole, constants.VirtualizationSystem)
	// logger.Infof("cpu info, model: %s, logical count: %d, physical count: %d",
	// 	constants.CpuModel, constants.CpuLogicalCount, constants.CpuPhysicalCount)
	// logger.Infof("disk info, total: %s, free: %s", utils.FormatBytes(int64(constants.DiskTotal)), utils.FormatBytes(int64(constants.DiskFree)))
	// logger.Infof("fs info, fs: %s, zfsmount: %s", constants.FsType, constants.DefaultZfsPrefixName)
	// logger.Infof("mem info, total: %s, free: %s", utils.FormatBytes(int64(constants.MemTotal)), utils.FormatBytes(int64(constants.MemFree)))
	// logger.Infof("cgroup info, cpu: %d, mem: %d", constants.CgroupCpuEnabled, constants.CgroupMemoryEnabled)

	return nil
}

type GreetingsTask struct {
	action.BaseAction
}

func (h *GreetingsTask) Execute(runtime connector.Runtime) error {
	_, err := runtime.GetRunner().Host.Cmd("echo 'Greetings, Terminus'", false, true)
	if err != nil {
		return err
	}

	return nil
}

type NodePreCheck struct {
	common.KubeAction
}

func (n *NodePreCheck) Execute(runtime connector.Runtime) error {
	var results = make(map[string]string)
	results["name"] = runtime.RemoteHost().GetName()
	for _, software := range baseSoftware {
		var (
			cmd string
		)

		switch software {
		case docker:
			cmd = "docker version --format '{{.Server.Version}}'"
		case containerd:
			cmd = "containerd --version | cut -d ' ' -f 3"
		default:
			cmd = fmt.Sprintf("which %s", software)
		}

		switch software {
		case sudo:
			// sudo skip sudo prefix
		default:
			cmd = connector.SudoPrefix(cmd)
		}

		res, err := runtime.GetRunner().Host.CmdExt(cmd, false, false)
		switch software {
		case showmount:
			software = nfs
		case rbd:
			software = ceph
		case glusterfs:
			software = glusterfs
		}
		if err != nil || strings.Contains(res, "not found") {
			results[software] = ""
		} else {
			// software in path
			if strings.Contains(res, "bin/") {
				results[software] = "y"
			} else {
				// get software version, e.g. docker, containerd, etc.
				results[software] = res
			}
		}
	}

	output, err := runtime.GetRunner().Host.CmdExt("date +\"%Z %H:%M:%S\"", false, false)
	if err != nil {
		results["time"] = ""
	} else {
		results["time"] = strings.TrimSpace(output)
	}

	host := runtime.RemoteHost()
	if res, ok := host.GetCache().Get(common.NodePreCheck); ok {
		m := res.(map[string]string)
		m = results
		host.GetCache().Set(common.NodePreCheck, m)
	} else {
		host.GetCache().Set(common.NodePreCheck, results)
	}
	return nil
}

type GetKubeConfig struct {
	common.KubeAction
}

func (g *GetKubeConfig) Execute(runtime connector.Runtime) error {
	var kubeConfigPath = "$HOME/.kube/config"
	if util.IsExist(kubeConfigPath) {
		return nil
	}

	if util.IsExist("/etc/kubernetes/admin.conf") {
		if _, err := runtime.GetRunner().Host.Cmd("mkdir -p $HOME/.kube", false, false); err != nil {
			return err
		}
		if _, err := runtime.GetRunner().Host.SudoCmd("cp /etc/kubernetes/admin.conf $HOME/.kube/config", false, false); err != nil {
			return err
		}
		// userId, err := runtime.GetRunner().Host.Cmd("echo $(id -u)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user id failed")
		// }

		// userGroupId, err := runtime.GetRunner().Host.Cmd("echo $(id -g)", false, false)
		// if err != nil {
		// 	return errors.Wrap(errors.WithStack(err), "get user group id failed")
		// }

		userId, err := runtime.GetRunner().Host.Cmd("echo $SUDO_UID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user id failed")
		}

		userGroupId, err := runtime.GetRunner().Host.Cmd("echo $SUDO_GID", false, false)
		if err != nil {
			return errors.Wrap(errors.WithStack(err), "get user group id failed")
		}

		chownKubeConfig := fmt.Sprintf("chown -R %s:%s $HOME/.kube", userId, userGroupId)
		if _, err := runtime.GetRunner().Host.SudoCmd(chownKubeConfig, false, false); err != nil {
			return errors.Wrap(errors.WithStack(err), "chown user kube config failed")
		}
	}

	return errors.New("kube config not found")
}

type GetAllNodesK8sVersion struct {
	common.KubeAction
}

func (g *GetAllNodesK8sVersion) Execute(runtime connector.Runtime) error {
	var nodeK8sVersion string
	kubeletVersionInfo, err := runtime.GetRunner().Host.SudoCmd("/usr/local/bin/kubelet --version", false, false)
	if err != nil {
		return errors.Wrap(err, "get current kubelet version failed")
	}
	nodeK8sVersion = strings.Split(kubeletVersionInfo, " ")[1]

	host := runtime.RemoteHost()
	if host.IsRole(common.Master) {
		apiserverVersion, err := runtime.GetRunner().Host.SudoCmd(
			"cat /etc/kubernetes/manifests/kube-apiserver.yaml | grep 'image:' | rev | cut -d ':' -f1 | rev",
			false, false)
		if err != nil {
			return errors.Wrap(err, "get current kube-apiserver version failed")
		}
		nodeK8sVersion = apiserverVersion
	}
	host.GetCache().Set(common.NodeK8sVersion, nodeK8sVersion)
	return nil
}

type CalculateMinK8sVersion struct {
	common.KubeAction
}

func (g *CalculateMinK8sVersion) Execute(runtime connector.Runtime) error {
	versionList := make([]*versionutil.Version, 0, len(runtime.GetHostsByRole(common.K8s)))
	for _, host := range runtime.GetHostsByRole(common.K8s) {
		version, ok := host.GetCache().GetMustString(common.NodeK8sVersion)
		if !ok {
			return errors.Errorf("get node %s Kubernetes version failed by host cache", host.GetName())
		}
		if versionObj, err := versionutil.ParseSemantic(version); err != nil {
			return errors.Wrap(err, "parse node version failed")
		} else {
			versionList = append(versionList, versionObj)
		}
	}

	minVersion := versionList[0]
	for _, version := range versionList {
		if !minVersion.LessThan(version) {
			minVersion = version
		}
	}
	g.PipelineCache.Set(common.K8sVersion, fmt.Sprintf("v%s", minVersion))
	return nil
}

type CheckDesiredK8sVersion struct {
	common.KubeAction
}

func (k *CheckDesiredK8sVersion) Execute(_ connector.Runtime) error {
	if ok := kubernetes.VersionSupport(k.KubeConf.Cluster.Kubernetes.Version); !ok {
		return errors.New(fmt.Sprintf("does not support upgrade to Kubernetes %s",
			k.KubeConf.Cluster.Kubernetes.Version))
	}
	k.PipelineCache.Set(common.DesiredK8sVersion, k.KubeConf.Cluster.Kubernetes.Version)
	return nil
}

type KsVersionCheck struct {
	common.KubeAction
}

func (k *KsVersionCheck) Execute(runtime connector.Runtime) error {
	ksVersionStr, err := runtime.GetRunner().Host.SudoCmd(
		"/usr/local/bin/kubectl get deploy -n  kubesphere-system ks-console -o jsonpath='{.metadata.labels.version}'",
		false, false)
	if err != nil {
		if k.KubeConf.Cluster.KubeSphere.Enabled {
			return errors.Wrap(err, "get kubeSphere version failed")
		} else {
			ksVersionStr = ""
		}
	}

	ccKsVersionStr, ccErr := runtime.GetRunner().Host.SudoCmd(
		"/usr/local/bin/kubectl get ClusterConfiguration ks-installer -n  kubesphere-system  -o jsonpath='{.metadata.labels.version}'",
		false, false)
	if ccErr == nil && ksVersionStr == "v3.1.0" {
		ksVersionStr = ccKsVersionStr
	}
	k.PipelineCache.Set(common.KubeSphereVersion, ksVersionStr)
	return nil
}

type DependencyCheck struct {
	common.KubeAction
}

func (d *DependencyCheck) Execute(_ connector.Runtime) error {
	currentKsVersion, ok := d.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return errors.New("get current KubeSphere version failed by pipeline cache")
	}
	desiredVersion := d.KubeConf.Cluster.KubeSphere.Version

	if d.KubeConf.Cluster.KubeSphere.Enabled {
		var version string
		if latest, ok := kubesphere.LatestRelease(desiredVersion); ok {
			version = latest.Version
		} else if ks, ok := kubesphere.DevRelease(desiredVersion); ok {
			version = ks.Version
		} else {
			r := regexp.MustCompile("v(\\d+\\.)?(\\d+\\.)?(\\*|\\d+)")
			version = r.FindString(desiredVersion)
		}

		ksInstaller, ok := kubesphere.VersionMap[version]
		if !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", desiredVersion))
		}

		if currentKsVersion != desiredVersion {
			if ok := ksInstaller.UpgradeSupport(currentKsVersion); !ok {
				return errors.New(fmt.Sprintf("Unsupported upgrade plan: %s to %s", currentKsVersion, desiredVersion))
			}
		}

		if ok := ksInstaller.K8sSupport(d.KubeConf.Cluster.Kubernetes.Version); !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				version, d.KubeConf.Cluster.Kubernetes.Version))
		}
	} else {
		ksInstaller, ok := kubesphere.VersionMap[currentKsVersion]
		if !ok {
			return errors.New(fmt.Sprintf("Unsupported version: %s", currentKsVersion))
		}

		if ok := ksInstaller.K8sSupport(d.KubeConf.Cluster.Kubernetes.Version); !ok {
			return errors.New(fmt.Sprintf("KubeSphere %s does not support running on Kubernetes %s",
				currentKsVersion, d.KubeConf.Cluster.Kubernetes.Version))
		}
	}
	return nil
}

type GetKubernetesNodesStatus struct {
	common.KubeAction
}

func (g *GetKubernetesNodesStatus) Execute(runtime connector.Runtime) error {
	nodeStatus, err := runtime.GetRunner().Host.SudoCmd("/usr/local/bin/kubectl get node -o wide", false, false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeStatus, nodeStatus)

	cri, err := runtime.GetRunner().Host.SudoCmd("/usr/local/bin/kubectl get node -o jsonpath=\"{.items[*].status.nodeInfo.containerRuntimeVersion}\"", false, false)
	if err != nil {
		return err
	}
	g.PipelineCache.Set(common.ClusterNodeCRIRuntimes, cri)
	return nil
}

type GetStorageKeyTask struct {
	common.KubeAction
}

func (t *GetStorageKeyTask) Execute(runtime connector.Runtime) error {
	kubectl, err := util.GetCommand(common.CommandKubectl)
	if err != nil {
		return fmt.Errorf("kubectl not found")
	}
	var storageAccessKey, storageSecretKey, storageToken, storageClusterId string
	var ctx, cancel = context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	if stdout, err := runtime.GetRunner().Host.CmdExtWithContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-ak}'", kubectl), false, false); err != nil {
		storageAccessKey = os.Getenv(common.ENV_AWS_ACCESS_KEY_ID_SETUP)
		if storageAccessKey == "" {
			logger.Errorf("storage access key not found")
		}
	} else {
		storageAccessKey = stdout
	}

	if stdout, err := runtime.GetRunner().Host.CmdExtWithContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-sk}'", kubectl), false, false); err != nil {
		storageSecretKey = os.Getenv(common.ENV_AWS_SECRET_ACCESS_KEY_SETUP)
		if storageSecretKey == "" {
			logger.Errorf("storage secret key not found")
		}
	} else {
		storageSecretKey = stdout
	}

	if stdout, err := runtime.GetRunner().Host.CmdExtWithContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.annotations.bytetrade\\.io/s3-sts}'", kubectl), false, false); err != nil {
		storageToken = os.Getenv(common.ENV_AWS_SESSION_TOKEN_SETUP)
		if storageToken == "" {
			logger.Errorf("storage token not found")
		}
	} else {
		storageToken = stdout
	}

	if stdout, err := runtime.GetRunner().Host.CmdExtWithContext(ctx, fmt.Sprintf("%s get terminus terminus -o jsonpath='{.metadata.labels.bytetrade\\.io/cluster-id}'", kubectl), false, false); err != nil {
		storageClusterId = os.Getenv(common.ENV_CLUSTER_ID)
		if storageClusterId == "" {
			logger.Errorf("storage cluster id not found")
		}
	} else {
		storageClusterId = stdout
	}

	t.PipelineCache.Set(common.CacheAccessKey, storageAccessKey)
	t.PipelineCache.Set(common.CacheSecretKey, storageSecretKey)
	t.PipelineCache.Set(common.CacheToken, storageToken)
	t.PipelineCache.Set(common.CacheClusterId, storageClusterId)

	logger.Infof("storage: cloud: %v, type: %s, bucket: %s, ak: %s, sk: %s, tk: %s, id: %s",
		t.KubeConf.Arg.IsCloudInstance, t.KubeConf.Arg.Storage.StorageType, t.KubeConf.Arg.Storage.StorageBucket,
		t.KubeConf.Arg.Storage.StorageAccessKey, t.KubeConf.Arg.Storage.StorageSecretKey, t.KubeConf.Arg.Storage.StorageToken, t.KubeConf.Arg.Storage.StorageClusterId)

	return nil
}

type RemoveChattr struct {
	common.KubeAction
}

func (t *RemoveChattr) Execute(runtime connector.Runtime) error {
	runtime.GetRunner().Host.SudoCmd("chattr -i /etc/hosts", false, true)
	runtime.GetRunner().Host.SudoCmd("chattr -i /etc/resolv.conf", false, true)
	return nil
}
