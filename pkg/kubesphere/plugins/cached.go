package plugins

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type CopyManifest struct {
	common.KubeAction
}

func (t *CopyManifest) Execute(runtime connector.Runtime) error {
	cachedDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir)
	maniDir := path.Join(runtime.GetRootDir(), cc.ImagesDir)
	if !util.IsExist(maniDir) {
		return fmt.Errorf("images manifest directory not exists !!!")
	}

	filepath.Walk(maniDir, func(pathx string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if strings.Contains(info.Name(), "tar") || strings.Contains(info.Name(), "tar.gz") {
			util.MoveFile(pathx, path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ImagesDir))
		} else {
			if err := util.CopyFile(pathx, fmt.Sprintf("%s/%s", cachedDir, info.Name())); err != nil {
				logger.Errorf("copy %s to %s failed: %v", pathx, cachedDir, err)
			}
		}

		return nil
	})
	return nil
}

type CachedBuilder struct {
	common.KubeAction
}

func (t *CachedBuilder) Execute(runtime connector.Runtime) error {
	cachedDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir)
	if !util.IsExist(cachedDir) {
		util.Mkdir(cachedDir)
	}

	cachedImageDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ImageCacheDir)
	if !util.IsExist(cachedImageDir) {
		util.Mkdir(cachedImageDir)
	}

	cachedPkgDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	if !util.IsExist(cachedPkgDir) {
		util.Mkdir(cachedPkgDir)
	}

	cachedBuildFilesDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.BuildFilesCacheDir)
	if !util.IsExist(cachedBuildFilesDir) {
		util.Mkdir(cachedBuildFilesDir)
	}

	return nil
}
