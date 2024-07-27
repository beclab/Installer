package packages

import (
	"fmt"
	"os/exec"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/cache"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
)

func DownloadInstallPackage(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache, sqlProvider storage.Provider) error {
	installPackage := files.NewKubeBinary("full-package", arch, version, path)
	installPackage.Provider = sqlProvider
	installPackage.WriteDownloadingLog = true
	downloadFiles := []*files.KubeBinary{installPackage}
	filesMap := make(map[string]*files.KubeBinary)
	for _, downloadFile := range downloadFiles {
		if err := downloadFile.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", downloadFile.FileName)
		}

		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, downloadFile.ID, downloadFile.Version)

		filesMap[downloadFile.ID] = downloadFile
		if util.IsExist(downloadFile.Path()) {
			if downloadFile.OverWrite {
				utils.DeleteFile(downloadFile.Path())
			} else {
				if err := downloadFile.SHA256Check(); err != nil {
					p := downloadFile.Path()
					_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
				} else {
					logger.Infof("%s %s is existed", common.LocalHost, downloadFile.FileName)
					continue
				}
			}
		}

		if err := downloadFile.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", downloadFile.ID, downloadFile.Url, err)
		}
	}

	return nil
}

func DownloadPackage(kubeConf *common.KubeConf, path, version, arch string, pipelineCache *cache.Cache) error {
	// debug
	file1 := files.NewKubeBinary("file1", arch, version, path)
	file2 := files.NewKubeBinary("file2", arch, version, path)
	file3 := files.NewKubeBinary("file3", arch, version, path)
	file4 := files.NewKubeBinary("full-package", arch, version, path)
	// file4 := files.NewKubeBinary("kubekey", arch, "0.1.20", path) // todo test kubekey

	downloadFiles := []*files.KubeBinary{file1, file2, file3, file4}

	filesMap := make(map[string]*files.KubeBinary)
	for _, downloadFile := range downloadFiles {
		if err := downloadFile.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", downloadFile.FileName)
		}

		filesMap[downloadFile.ID] = downloadFile
		if util.IsExist(downloadFile.Path()) {
			// download it again if it's incorrect
			if err := downloadFile.SHA256Check(); err != nil {
				p := downloadFile.Path()
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Infof(common.LocalHost, "%s is existed", downloadFile.ID)
				continue
			}
		}

		if err := downloadFile.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", downloadFile.ID, downloadFile.Url, err)
		}
	}

	return nil
}
