/*
 Copyright 2022 The KubeSphere Authors.

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

// RegistryPackageDownloadHTTP defines the kubernetes' binaries that need to be downloaded in advance and downloads them.
func RegistryPackageDownloadHTTP(kubeConf *common.KubeConf, path, arch string, pipelineCache *cache.Cache) error {
	var binaries []*files.KubeBinary

	switch kubeConf.Cluster.Registry.Type {
	case common.Harbor:
		// TODO: Harbor only supports amd64, so there is no need to consider other architectures at present.
		harbor := files.NewKubeBinary("harbor", arch, kubekeyapiv1alpha2.DefaultHarborVersion, path)
		compose := files.NewKubeBinary("compose", arch, kubekeyapiv1alpha2.DefaultDockerComposeVersion, path)
		docker := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path)
		binaries = []*files.KubeBinary{harbor, docker, compose}
	default:
		registry := files.NewKubeBinary("registry", arch, kubekeyapiv1alpha2.DefaultRegistryVersion, path)
		binaries = []*files.KubeBinary{registry}
	}

	binariesMap := make(map[string]*files.KubeBinary)
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Infof("%s downloading %s %s %s  ...", common.LocalHost, arch, binary.ID, binary.Version)
		binariesMap[binary.ID] = binary
		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				continue
			}
		}

		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	pipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)
	return nil
}

func RegistryBinariesDownload(manifest *common.ArtifactManifest, path, arch string) error {

	m := manifest.Spec
	binaries := make([]*files.KubeBinary, 0, 0)

	if m.Components.DockerRegistry.Version != "" {
		registry := files.NewKubeBinary("registry", arch, kubekeyapiv1alpha2.DefaultRegistryVersion, path)
		binaries = append(binaries, registry)
	}

	if m.Components.Harbor.Version != "" {
		harbor := files.NewKubeBinary("harbor", arch, kubekeyapiv1alpha2.DefaultHarborVersion, path)
		// TODO: Harbor only supports amd64, so there is no need to consider other architectures at present.
		if arch == "amd64" {
			binaries = append(binaries, harbor)
		}
	}

	if m.Components.DockerCompose.Version != "" {
		compose := files.NewKubeBinary("compose", arch, kubekeyapiv1alpha2.DefaultDockerComposeVersion, path)
		// TODO: Harbor only supports amd64, so there is no need to consider other architectures at present. docker-compose is required only if harbor is installed.
		containerManager := files.NewKubeBinary("docker", arch, kubekeyapiv1alpha2.DefaultDockerVersion, path)
		if arch == "amd64" {
			binaries = append(binaries, compose)
			binaries = append(binaries, containerManager)
		}
	}

	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)

		if util.IsExist(binary.Path()) {
			// download it again if it's incorrect
			if err := binary.SHA256Check(); err != nil {
				p := binary.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
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
