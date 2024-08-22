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

package binaries

import (
	"fmt"
	"os/exec"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"github.com/pkg/errors"
)

// K3sFilesDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func K3sFilesDownloadHTTP(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache) error {

	etcd := files.NewKubeBinary("etcd", arch, kubekeyapiv1alpha2.DefaultEtcdVersion, path)
	kubecni := files.NewKubeBinary("kubecni", arch, kubekeyapiv1alpha2.DefaultCniVersion, path)
	helm := files.NewKubeBinary("helm", arch, kubekeyapiv1alpha2.DefaultHelmVersion, path)
	k3s := files.NewKubeBinary("k3s", arch, version, path)

	crictl := files.NewKubeBinary("crictl", arch, kubekeyapiv1alpha2.DefaultCrictlVersion, path)
	containerd := files.NewKubeBinary("containerd", arch, kubekeyapiv1alpha2.DefaultContainerdVersion, path)
	runc := files.NewKubeBinary("runc", arch, kubekeyapiv1alpha2.DefaultRuncVersion, path)

	binaries := []*files.KubeBinary{k3s, helm, kubecni, etcd, crictl, containerd, runc}
	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		binariesMap[binary.ID] = binary
		var exists = util.IsExist(binary.Path())
		if exists {
			p := binary.Path()
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Infof("%s exists", binary.FileName)
				if binary.ID == "helm" {
					util.CopyFile(fmt.Sprintf("%s/helm", binary.BaseDir), "/usr/local/bin/helm")
					_ = exec.Command("/bin/sh", "-c", "chmod +x /usr/local/bin/helm").Run()
				}
				continue
			}
		}

		if !exists && binary.ID == "helm" {
			if util.IsExist(fmt.Sprintf("%s/helm", binary.BaseDir)) {
				logger.Infof("%s exists", binary.FileName)
				util.CopyFile(fmt.Sprintf("%s/helm", binary.BaseDir), "/usr/local/bin/helm")
				_ = exec.Command("/bin/sh", "-c", "chmod +x /usr/local/bin/helm").Run()
				continue
			}
		}

		if !exists || binary.OverWrite {
			logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)
			if err := binary.Download(); err != nil {
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
			}
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}

func K3sArtifactBinariesDownload(manifest *common.ArtifactManifest, path, arch, version string) error {
	m := manifest.Spec

	etcd := files.NewKubeBinary("etcd", arch, m.Components.ETCD.Version, path)
	kubecni := files.NewKubeBinary("kubecni", arch, m.Components.CNI.Version, path)
	helm := files.NewKubeBinary("helm", arch, m.Components.Helm.Version, path)
	k3s := files.NewKubeBinary("k3s", arch, version, path)
	crictl := files.NewKubeBinary("crictl", arch, m.Components.Crictl.Version, path)
	binaries := []*files.KubeBinary{k3s, helm, kubecni, etcd}

	dockerArr := make([]*files.KubeBinary, 0, 0)
	dockerVersionMap := make(map[string]struct{})
	for _, c := range m.Components.ContainerRuntimes {
		var dockerVersion string
		if c.Type == common.Docker {
			dockerVersion = c.Version
		} else {
			dockerVersion = kubekeyapiv1alpha2.DefaultDockerVersion
		}
		if _, ok := dockerVersionMap[dockerVersion]; !ok {
			dockerVersionMap[dockerVersion] = struct{}{}
			docker := files.NewKubeBinary("docker", arch, dockerVersion, path)
			dockerArr = append(dockerArr, docker)
		}
	}

	binaries = append(binaries, dockerArr...)
	if m.Components.Crictl.Version != "" {
		binaries = append(binaries, crictl)
	}

	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Debugf("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)

		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", binary.Path())).Run()
			} else {
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	return nil
}
