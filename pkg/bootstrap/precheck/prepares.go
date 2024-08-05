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
	"fmt"
	"net"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/prepare"
	"github.com/pkg/errors"
)

type LocalIpCheck struct {
	prepare.BasePrepare
}

func (p *LocalIpCheck) PreCheck(runtime connector.Runtime) (bool, error) {
	var localIp = constants.LocalIp
	ip := net.ParseIP(localIp)
	if ip == nil {
		return false, fmt.Errorf("invalid local ip %s", localIp)
	}

	if ip4 := ip.To4(); ip4 == nil {
		return false, fmt.Errorf("invalid local ip %s", localIp)
	}

	switch localIp {
	case "172.17.0.1", "127.0.0.1", "127.0.1.1":
		return false, fmt.Errorf("invalid local ip %s", localIp)
	default:
	}
	return true, nil
}

type OsSupportCheck struct {
	prepare.BasePrepare
}

func (p *OsSupportCheck) PreCheck(runtime connector.Runtime) (bool, error) {
	switch constants.OsType {
	case common.Linux:
		switch constants.OsPlatform {
		case common.Ubuntu:
			if strings.HasPrefix(constants.OsVersion, "20.") || strings.HasPrefix(constants.OsVersion, "22.") || strings.HasPrefix(constants.OsVersion, "24.") {
				return true, nil
			}
			return false, fmt.Errorf("os %s version %s not support", constants.OsPlatform, constants.OsVersion)
		case common.Debian:
			if strings.HasPrefix(constants.OsVersion, "11") || strings.HasPrefix(constants.OsVersion, "12") {
				return true, nil
			}
			return false, fmt.Errorf("os %s version %s not support", constants.OsPlatform, constants.OsVersion)
		default:
			return false, fmt.Errorf("platform %s not support", constants.OsPlatform)
		}
	default:
		return false, fmt.Errorf("os %s not support", constants.OsType)
	}
}

type KubeSphereExist struct {
	common.KubePrepare
}

func (k *KubeSphereExist) PreCheck(runtime connector.Runtime) (bool, error) {
	currentKsVersion, ok := k.PipelineCache.GetMustString(common.KubeSphereVersion)
	if !ok {
		return false, errors.New("get current KubeSphere version failed by pipeline cache")
	}
	if currentKsVersion != "" {
		return true, nil
	}
	return false, nil
}

type KubeExist struct {
	common.KubePrepare
}

func (k *KubeExist) PreCheck(runtime connector.Runtime) (bool, error) {
	if constants.InstalledKubeVersion == "" {
		return false, nil
	}
	return true, nil
}
