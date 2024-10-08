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
	"time"

	"bytetrade.io/web3os/installer/pkg/binaries"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/module"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/manifest"
)

type RemoveChattrModule struct {
	common.KubeModule
}

func (m *RemoveChattrModule) Init() {
	m.Name = "RemoveChattr"

	removeChattr := &task.RemoteTask{
		Name:     "RemoveChattr",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(RemoveChattr),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		removeChattr,
	}
}

type GetStorageKeyModule struct {
	common.KubeModule
}

func (m *GetStorageKeyModule) Init() {
	m.Name = "GetStorage"

	getStorageKeyTask := &task.RemoteTask{
		Name:     "GetStorageKey",
		Hosts:    m.Runtime.GetHostsByRole(common.Master),
		Action:   new(GetStorageKeyTask),
		Parallel: false,
		Retry:    1,
	}

	m.Tasks = []task.Interface{
		getStorageKeyTask,
	}
}

type GetKubeVersionModule struct {
	module.BaseTaskModule
}

type GetSysInfoModel struct {
	module.BaseTaskModule
}

func (m *GetSysInfoModel) Init() {
	m.Name = "GetInfo"

	getSysInfoTask := &task.LocalTask{
		Name:   "GetSysInfo",
		Action: new(GetSysInfoTask),
	}

	m.Tasks = []task.Interface{
		getSysInfoTask,
	}
}

type PreCheckOsModule struct {
	common.KubeModule
	manifest.ManifestModule
}

func (m *PreCheckOsModule) Init() {
	m.Name = "PreCheckOs"

	patchAppArmor := &task.RemoteTask{
		Name:  "PatchAppArmor",
		Hosts: m.Runtime.GetAllHosts(),
		Prepare: &prepare.PrepareCollection{
			&binaries.Ubuntu24AppArmorCheck{},
		},
		Action: &binaries.InstallAppArmorTask{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest},
		},
		Parallel: false,
		Retry:    0,
	}

	raspbianCheck := &task.RemoteTask{
		Name:     "RaspbianCheck",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(RaspbianCheckTask),
		Parallel: false,
		Retry:    0,
	}

	correctHostname := &task.RemoteTask{
		Name:     "CorrectHostname",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(CorrectHostname),
		Parallel: false,
		Retry:    0,
	}

	disableDNS := &task.RemoteTask{
		Name:     "DisableLocalDNS",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(DisableLocalDNSTask),
		Parallel: false,
		Retry:    0,
	}

	m.Tasks = []task.Interface{
		patchAppArmor,
		raspbianCheck,
		correctHostname,
		disableDNS,
	}
}

type GreetingsModule struct {
	module.BaseTaskModule
}

func (h *GreetingsModule) Init() {
	h.Name = "GreetingsModule"

	var timeout int64

	for _, v := range h.Runtime.GetAllHosts() {
		timeout += v.GetTimeout()
	}

	hello := &task.RemoteTask{
		Name:     "Greetings",
		Hosts:    h.Runtime.GetAllHosts(),
		Action:   new(GreetingsTask),
		Parallel: false,
		Timeout:  time.Duration(timeout) * time.Second,
	}

	h.Tasks = []task.Interface{
		hello,
	}
}

type NodePreCheckModule struct {
	common.KubeModule
	Skip bool
}

func (n *NodePreCheckModule) IsSkip() bool {
	return n.Skip
}

func (n *NodePreCheckModule) Init() {
	n.Name = "NodePreCheckModule"
	n.Desc = "Do pre-check on cluster nodes"

	preCheck := &task.RemoteTask{
		Name:  "NodePreCheck",
		Desc:  "A pre-check on nodes",
		Hosts: n.Runtime.GetAllHosts(),
		//Prepare: &prepare.FastPrepare{
		//	Inject: func(runtime connector.Runtime) (bool, error) {
		//		if len(n.Runtime.GetHostsByRole(common.ETCD))%2 == 0 {
		//			logger.Error("The number of etcd is even. Please configure it to be odd.")
		//			return false, errors.New("the number of etcd is even")
		//		}
		//		return true, nil
		//	}},
		Action:   new(NodePreCheck),
		Parallel: true,
	}

	n.Tasks = []task.Interface{
		preCheck,
	}
}

type ClusterPreCheckModule struct {
	common.KubeModule
}

func (c *ClusterPreCheckModule) Init() {
	c.Name = "ClusterPreCheckModule"
	c.Desc = "Do pre-check on cluster"

	getKubeConfig := &task.RemoteTask{
		Name:     "GetKubeConfig",
		Desc:     "Get KubeConfig file",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GetKubeConfig),
		Parallel: true,
	}

	getAllNodesK8sVersion := &task.RemoteTask{
		Name:     "GetAllNodesK8sVersion",
		Desc:     "Get all nodes Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.K8s),
		Action:   new(GetAllNodesK8sVersion),
		Parallel: true,
	}

	calculateMinK8sVersion := &task.RemoteTask{
		Name:     "CalculateMinK8sVersion",
		Desc:     "Calculate min Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CalculateMinK8sVersion),
		Parallel: true,
	}

	checkDesiredK8sVersion := &task.RemoteTask{
		Name:     "CheckDesiredK8sVersion",
		Desc:     "Check desired Kubernetes version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(CheckDesiredK8sVersion),
		Parallel: true,
	}

	ksVersionCheck := &task.RemoteTask{
		Name:     "KsVersionCheck",
		Desc:     "Check KubeSphere version",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(KsVersionCheck),
		Parallel: true,
	}

	dependencyCheck := &task.RemoteTask{
		Name:  "DependencyCheck",
		Desc:  "Check dependency matrix for KubeSphere and Kubernetes",
		Hosts: c.Runtime.GetHostsByRole(common.Master),
		Prepare: &prepare.PrepareCollection{
			new(common.OnlyFirstMaster),
			new(KubeSphereExist),
		},
		Action:   new(DependencyCheck),
		Parallel: true,
	}

	getKubernetesNodesStatus := &task.RemoteTask{
		Name:     "GetKubernetesNodesStatus",
		Desc:     "Get kubernetes nodes status",
		Hosts:    c.Runtime.GetHostsByRole(common.Master),
		Prepare:  new(common.OnlyFirstMaster),
		Action:   new(GetKubernetesNodesStatus),
		Parallel: true,
	}

	c.Tasks = []task.Interface{
		getKubeConfig,
		getAllNodesK8sVersion,
		calculateMinK8sVersion,
		checkDesiredK8sVersion,
		ksVersionCheck,
		dependencyCheck,
		getKubernetesNodesStatus,
	}
}
