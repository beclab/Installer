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

package packages

import (
	"fmt"
	"path"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	corecommon "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type PackageDownload struct {
	common.KubeAction
}

func (d *PackageDownload) Execute(runtime connector.Runtime) error {
	provider := runtime.GetStorage()
	if provider == nil {
	}
	// if err := DownloadInstallPackage(d.KubeConf, runtime.GetPackageDir(), "0.0.1", kubekeyapiv1alpha2.DefaultArch, d.PipelineCache, provider); err != nil {
	// 	return err
	// }

	if err := DownloadPackage(d.KubeConf, runtime.GetPackageDir(), "0.0.1", kubekeyapiv1alpha2.DefaultArch, d.PipelineCache); err != nil {
		return err
	}

	return nil
}

type PackageUntar struct {
	common.KubeAction
}

func (a *PackageUntar) Execute(runtime connector.Runtime) error {

	var pkgFile = fmt.Sprintf("%s/install-wizard-full.tar.gz", runtime.GetPackageDir())
	if ok := util.IsExist(pkgFile); !ok {
		return fmt.Errorf("package %s not exist", pkgFile)
	}

	var p = path.Join(runtime.GetPackageDir(), corecommon.InstallDir)
	// ./packages/
	if err := util.RemoveDir(p); err != nil {
		return fmt.Errorf("remove %s failed %v", p, err)
	}

	if err := util.Mkdir(p); err != nil {
		return fmt.Errorf("mkdir %s failed %v", p, err)
	}

	var provider = runtime.GetStorage()
	if err := provider.SaveInstallLog(fmt.Sprintf("decompressing %s, please wait", "install-wizard-full.tar.gz"), corecommon.StateDownload, int64(2*10000/corecommon.DefaultInstallSteps)); err != nil {
		logger.Errorf("save download log failed %v", err)
	}
	if err := util.Untar(pkgFile, p); err != nil {
		return fmt.Errorf("untar %s failed %v", pkgFile, err)
	}
	logger.Debugf("decompression %s success", pkgFile)
	return nil
}
