package download

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type PackageDownload struct {
	common.KubeAction
	Manifest string
	BaseDir  string
}

type CheckDownload struct {
	common.KubeAction
	Manifest string
	BaseDir  string
}

func (d *PackageDownload) Execute(runtime connector.Runtime) error {
	if d.Manifest == "" {
		return errors.New("manifest path is empty")
	}

	var baseDir = d.BaseDir
	var systemInfo = runtime.GetSystemInfo()

	if systemInfo.IsWsl() {
		var wslPackageDir = d.KubeConf.Arg.GetWslUserPath()
		if wslPackageDir != "" {
			baseDir = fmt.Sprintf("%s/.terminus", wslPackageDir)
		}
	}

	if data, err := os.ReadFile(d.Manifest); err != nil {
		logger.Fatal("unable to read manifest, ", err)
	} else {
		scanner := bufio.NewScanner(bytes.NewReader(data))

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" && strings.HasPrefix(line, "#") {
				continue
			}

			item := must(manifest.ReadItem(line))

			if !must(isRealExists(runtime, item, baseDir)) {
				err := d.downloadItem(runtime, item, baseDir)
				if err != nil {
					logger.Fatal(err)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			logger.Fatal("reading mainfest err:", err)
		}
	}

	return nil
}

func (d *CheckDownload) Execute(runtime connector.Runtime) error {
	if d.Manifest == "" {
		return errors.New("manifest path is empty")
	}

	if data, err := os.ReadFile(d.Manifest); err != nil {
		logger.Fatal("unable to read manifest, ", err)
	} else {
		scanner := bufio.NewScanner(bytes.NewReader(data))

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" && strings.HasPrefix(line, "#") {
				continue
			}

			item := must(manifest.ReadItem(line))

			if !must(isRealExists(runtime, item, d.BaseDir)) {
				name := item.Filename
				if item.ImageName != "" {
					name = item.ImageName
				}
				logger.Fatal("%s not found in the pre-download path", name)
			}
		}
	}

	logger.Info("suceess to check download")
	return nil

}

// if the file exists and the checksum passed
func isRealExists(runtime connector.Runtime, item *manifest.ManifestItem, baseDir string) (bool, error) {
	arch := runtime.GetSystemInfo().GetOsArch()
	targetPath := getDownloadTargetPath(item, baseDir)
	exists := runtime.GetRunner().Host.FileExist(targetPath)
	if !exists {
		return false, nil
	}

	checksum := utils.LocalMd5Sum(targetPath)
	// FIXME: run in remote
	return checksum == item.GetItemUrlForHost(arch).Checksum, nil
}

func (d *PackageDownload) downloadItem(runtime connector.Runtime, item *manifest.ManifestItem, baseDir string) error {
	arch := runtime.GetSystemInfo().GetOsArch()
	url := item.GetItemUrlForHost(arch)

	component := new(files.KubeBinary)
	component.ID = item.Filename
	component.Arch = runtime.GetSystemInfo().GetOsArch()
	component.BaseDir = getDownloadTargetBasePath(item, baseDir)
	component.Url = url.Url
	component.FileName = item.Filename
	component.CheckMd5Sum = true
	component.Md5sum = url.Checksum

	downloadPath := component.Path()
	if utils.IsExist(downloadPath) {
		_, _ = runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf %s", downloadPath), false, false)
	}

	if !utils.IsExist(component.BaseDir) {
		if err := component.CreateBaseDir(); err != nil {
			return err
		}
	}

	if err := component.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", component.ID, component.Url, err)
	}

	return nil
}

func getDownloadTargetPath(item *manifest.ManifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s/%s", baseDir, item.Path, item.Filename)
}

func getDownloadTargetBasePath(item *manifest.ManifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s", baseDir, item.Path)
}

func must[T any](r T, e error) T {
	if e != nil {
		logger.Fatal(e)
	}

	return r
}
