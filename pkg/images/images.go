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

package images

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

const (
	cnRegistry          = "registry.cn-beijing.aliyuncs.com"
	cnNamespaceOverride = "kubesphereio"
)

// Image defines image's info.
type Image struct {
	RepoAddr          string
	Namespace         string
	NamespaceOverride string
	Repo              string
	Tag               string
	Group             string
	Enable            bool
}

// Images contains a list of Image
type Images struct {
	Images []Image
}

// ImageName is used to generate image's full name.
func (image Image) ImageName() string {
	return fmt.Sprintf("%s:%s", image.ImageRepo(), image.Tag)
}

// ImageRepo is used to generate image's repo address.
func (image Image) ImageRepo() string {
	var prefix string

	if os.Getenv("KKZONE") == "cn" {
		if image.RepoAddr == "" || image.RepoAddr == cnRegistry {
			image.RepoAddr = cnRegistry
			image.NamespaceOverride = cnNamespaceOverride
		}
	}

	if image.RepoAddr == "" {
		if image.Namespace == "" {
			prefix = ""
		} else {
			prefix = fmt.Sprintf("%s/", image.Namespace)
		}
	} else {
		if image.NamespaceOverride == "" {
			if image.Namespace == "" {
				prefix = fmt.Sprintf("%s/library/", image.RepoAddr)
			} else {
				prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.Namespace)
			}
		} else {
			prefix = fmt.Sprintf("%s/%s/", image.RepoAddr, image.NamespaceOverride)
		}
	}

	return fmt.Sprintf("%s%s", prefix, image.Repo)
}

// PullImages is used to pull images in the list of Image.
func (images *Images) PullImages(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	pullCmd := "docker"
	switch kubeConf.Cluster.Kubernetes.ContainerManager {
	case "crio":
		pullCmd = "crictl"
	case "containerd":
		pullCmd = "crictl"
	case "isula":
		pullCmd = "isula"
	default:
		pullCmd = "docker"
	}

	host := runtime.RemoteHost()

	// TODO Will back up image files locally in the future
	var imagePath, err = utils.GetRealPath(path.Join(runtime.GetRootDir(), "images"))
	if err != nil {
		return err
	}
	logger.Debugf("images path: %s", imagePath)
	if util.IsExist(imagePath) {
		var manifestPath = path.Join(imagePath, common.ManifestImageList)
		realManifestPath, err := utils.GetRealPath(manifestPath)
		if err != nil {
			logger.Errorf("get manifest file error %v, path: %s", err, manifestPath)
			return err
		}
		mf, _ := readImageManifest(realManifestPath)
		if mf == nil {
			logger.Debugf("image manifest not found, skip")
			return nil
		}

		for _, imageRepoTag := range mf {
			var inspect = fmt.Sprintf("%s inspecti -q %s", pullCmd, imageRepoTag)
			_, err := runtime.GetRunner().SudoCmdExt(inspect, false, false)
			if err == nil {
				continue
			}
			var imageFileNamePrefix = utils.MD5(imageRepoTag)
			var found string
			filepath.Walk(imagePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				if !strings.HasPrefix(info.Name(), imageFileNamePrefix) {
					return nil
				}

				realImagePath, err := utils.GetRealPath(path)
				if err != nil {
					return nil
				}

				if !HasSuffixI(info.Name(), ".tar.gz", ".tgz", ".tar") {
					// logger.Debugf("image file %s name is not standardized. Supported file formats include .tar.gz, .tgz, .tar", info.Name())
					return nil
				}

				found = info.Name()

				var importCmd = " ctr -n k8s.io images import "
				if HasSuffixI(realImagePath, ".tar.gz", ".tgz") {
					importCmd = fmt.Sprintf("gunzip -c %s | %s -", realImagePath, importCmd)
				} else {
					importCmd = fmt.Sprintf("%s %s", importCmd, realImagePath)
				}

				var ts = time.Now()
				if _, err = runtime.GetRunner().SudoCmdExt(importCmd, false, false); err != nil {
					return fmt.Errorf("import image %s %s failed", imageRepoTag, realImagePath)
				}

				fmt.Printf("unpacking %s done(%s)\n", imageRepoTag, util.ShortDur(time.Since(ts)))
				return filepath.SkipDir
			})

			if found == "" {
				fmt.Printf("image %s file %s.* not found\n", imageRepoTag, imageFileNamePrefix)
				logger.Infof("image %s file %s.* not found ", imageRepoTag, imageFileNamePrefix)
			}
		}
	}

	for _, image := range images.Images {
		switch {
		case host.IsRole(common.Master) && image.Group == kubekeyapiv1alpha2.Master && image.Enable,
			host.IsRole(common.Worker) && image.Group == kubekeyapiv1alpha2.Worker && image.Enable,
			(host.IsRole(common.Master) || host.IsRole(common.Worker)) && image.Group == kubekeyapiv1alpha2.K8s && image.Enable,
			host.IsRole(common.ETCD) && image.Group == kubekeyapiv1alpha2.Etcd && image.Enable:

			if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("%s inspecti -q %s", pullCmd, image.ImageName()), false, false); err == nil {
				logger.Infof("%s pull image %s exists", pullCmd, image.ImageName())
				continue
			}

			// fmt.Printf("%s downloading image %s\n", pullCmd, image.ImageName())
			logger.Debugf("%s downloading image: %s - %s", host.GetName(), image.ImageName(), runtime.RemoteHost().GetName())
			if _, err := runtime.GetRunner().SudoCmd(fmt.Sprintf("%s pull %s", pullCmd, image.ImageName()), false, false); err != nil {
				return errors.Wrap(err, "pull image failed")
			}
		default:
			continue
		}

	}
	return nil
}

type LocalImage struct {
	Filename string
}

type LocalImages []LocalImage

func (i LocalImages) LoadImages(runtime connector.Runtime, kubeConf *common.KubeConf) error {
	loadCmd := "docker"

	host := runtime.RemoteHost()

	// todo
	// var decompressDir = path.Join(common.TmpDir, "images")
	// if !util.IsExist(decompressDir) {
	// 	util.Mkdir(decompressDir)
	// }

	// for _, image := range i {
	// 	var dst = strings.ReplaceAll(image.Filename, ".gz", "")
	// 	var dstFile = filepath.Base(dst)
	// 	var dstName = path.Join(decompressDir, dstFile)
	// 	var cmd = fmt.Sprintf("gunzip -c %s > %s", image.Filename, dstName)
	// 	if _, err := runtime.GetRunner().SudoCmd(cmd, false, false); err != nil {
	// 		logger.Infof("gunzip image %s failed %v", err)
	// 		return err
	// 	}
	// 	logger.Debugf("gunzip %s successed", image.Filename)
	// 	image.Filename = dstName
	// 	time.Sleep(1 * time.Second)
	// }

	retry := func(f func() error, times int) (err error) {
		for i := 0; i < times; i++ {
			err = f()
			if err == nil {
				return nil
			}
			var dur = 5 + (i+1)*10
			logger.Debugf("load image %s failed, wait for %d seconds(%d times)", err, dur, i+1)
			if (i + 1) < times {
				time.Sleep(time.Duration(dur) * time.Second)
			}
		}

		return
	}

	for _, image := range i {
		switch {
		case host.IsRole(common.Master):
			// logger.Debugf("%s preloading image: %s", host.GetName(), image.Filename)
			start := time.Now()
			fileName := filepath.Base(image.Filename)
			// fileName = strings.ReplaceAll(fileName, ".gz", "")
			// fmt.Println(">>> ", fileName, HasSuffixI(image.Filename, ".tar.gz", ".tgz"))
			if HasSuffixI(image.Filename, ".tar.gz", ".tgz") {
				switch kubeConf.Cluster.Kubernetes.ContainerManager {
				case "crio":
					loadCmd = "ctr" // BUG
				case "containerd":
					loadCmd = "ctr -n k8s.io images import -"
				case "isula":
					loadCmd = "isula"
				default:
					loadCmd = "docker load"
				}

				// continue if load image error
				if err := retry(func() error {
					logger.Debugf("preloading image: %s", fileName)
					if stdout, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("env PATH=$PATH gunzip -c %s | %s", image.Filename, loadCmd), false, false); err != nil {
						return fmt.Errorf("%s", fileName)
					} else {
						logger.Debugf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
						// fmt.Printf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
					}
					return nil
				}, 5); err != nil {
					// logger.Errorf("load %s failed: %v in %s", fileName, err, time.Since(start))
					// os.Exit(1)
					// return err
					return fmt.Errorf("%s", fileName)
				}
			} else if HasSuffixI(image.Filename, ".tar") {
				switch kubeConf.Cluster.Kubernetes.ContainerManager {
				case "crio":
					loadCmd = "ctr" // BUG
				case "containerd":
					loadCmd = "ctr -n k8s.io images import"
				case "isula":
					loadCmd = "isula"
				default:
					loadCmd = "docker load -i"
				}

				if err := retry(func() error {
					logger.Debugf("preloading image: %s", fileName)
					if stdout, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("env PATH=$PATH %s %s", loadCmd, image.Filename), false, false); err != nil {
						return fmt.Errorf("%s", fileName)
					} else {
						logger.Debugf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
						// fmt.Printf("%s in %s\n", formatLoadImageRes(stdout, fileName), time.Since(start))
					}

					return nil
				}, 5); err != nil {
					return fmt.Errorf("%s", fileName)
				}
			} else {
				logger.Debugf("invalid image file name %s, skip ...", image.Filename)
				return nil
			}
		default:
			continue
		}

	}
	return nil

}

func formatLoadImageRes(str string, fileName string) string {
	if strings.Contains(str, "(sha256:") {
		str = strings.Split(str, "(sha256:")[0]
	} else {
		return fmt.Sprintf("%s %s", str, fileName)
	}
	return fmt.Sprintf("%s (%s)...done ", str, fileName)
}

func HasSuffixI(s string, suffixes ...string) bool {
	s = strings.ToLower(s)
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, strings.ToLower(suffix)) {
			return true
		}
	}
	return false
}

func readImageManifest(mfPath string) ([]string, error) {
	if !util.IsExist(mfPath) {
		return nil, nil
	}

	file, err := os.Open(mfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var res []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		res = append(res, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return res, nil
}
