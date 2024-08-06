package plugins

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type CachedManifest struct {
	common.KubeAction
}

func (t *CachedManifest) Execute(runtime connector.Runtime) error {
	cachedDir := path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir)
	if !util.IsExist(cachedDir) {
		util.Mkdir(cachedDir)
	}

	maniDir := path.Join(runtime.GetRootDir(), fmt.Sprintf(".%s", cc.ManifestDir))
	if !util.IsExist(maniDir) {
		panic(".manifest not found")
	}

	filepath.Walk(maniDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if err := util.CopyFile(path, fmt.Sprintf("%s/%s", cachedDir, info.Name())); err != nil {
			logger.Errorf("copy %s to %s failed: %v", path, cachedDir, err)
		}
		return nil
	})

	return nil
}
