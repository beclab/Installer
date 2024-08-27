package download

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
)

type PackageDownload struct {
	common.KubeAction
	Manifest string
}

type CheckDownload struct {
	common.KubeAction
	Manifest string
}

type fileUrl struct {
	Url      string
	Checksum string // md5 checksum
}

type itemUrl struct {
	AMD64 fileUrl
	ARM64 fileUrl
}

type manifestItem struct {
	Filename  string
	Path      string
	Type      string
	URL       itemUrl
	ImageName string
}

func (d *PackageDownload) Execute(runtime connector.Runtime) error {
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

			item := must(readItem(line))
			baseDir := runtime.GetHomeDir() + "/.terminus"

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

			item := must(readItem(line))
			baseDir := runtime.GetHomeDir() + "/.terminus"

			if !must(isRealExists(runtime, item, baseDir)) {
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

func readItem(line string) (*manifestItem, error) {
	token := strings.Split(line, ",")
	if len(token) < 7 {
		return nil, errors.New("invalid format")
	}

	item := &manifestItem{
		Filename: token[0],
		Path:     token[1],
		Type:     token[2],
		URL: itemUrl{
			AMD64: fileUrl{
				Url:      token[3],
				Checksum: token[4],
			},
			ARM64: fileUrl{
				Url:      token[5],
				Checksum: token[6],
			},
		},
	}
	if strings.HasPrefix(token[2], "images.") && len(token) > 7 {
		item.ImageName = token[7]
	}

	return item, nil
}

// if the file exists and the checksum passed
func isRealExists(runtime connector.Runtime, item *manifestItem, baseDir string) (bool, error) {
	targetPath := getDownloadTargetPath(item, baseDir)
	if exists, err := runtime.GetRunner().FileExist(targetPath); err != nil {
		return false, err
	} else if !exists {
		return false, nil
	}

	checksum := utils.LocalMd5Sum(targetPath)
	// FIXME: run in remote
	return checksum == getItemUrlForHost(item).Checksum, nil
}

func (d *PackageDownload) downloadItem(runtime connector.Runtime, item *manifestItem, baseDir string) error {
	url := getItemUrlForHost(item)

	component := new(files.KubeBinary)
	component.ID = item.Filename
	component.Arch = constants.OsArch
	component.BaseDir = getDownloadTargetBasePath(item, baseDir)
	component.Url = url.Url

	downloadPath := component.Path()
	if utils.IsExist(downloadPath) {
		_, _ = runtime.GetRunner().SudoCmdExt(fmt.Sprintf("rm -rf %s", downloadPath), false, false)
	}

	if !utils.IsExist(component.BaseDir) {
		if err := component.CreateBaseDir(); err != nil {
			return err
		}
	}

	if err := component.Download(); err != nil {
		return fmt.Errorf("Failed to download %s binary: %s error: %w ", component.ID, component.Url, err)
	}

	checksum := utils.LocalMd5Sum(component.Path())
	if checksum != url.Checksum {
		return fmt.Errorf("Failed to download %s binary: %s error: checksum failed ", component.ID, component.Url)
	}

	return nil
}

func getItemUrlForHost(item *manifestItem) *fileUrl {
	switch constants.OsArch {
	case "arm64":
		return &item.URL.ARM64
	}

	return &item.URL.AMD64
}

func getDownloadTargetPath(item *manifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s/%s", baseDir, item.Path, item.Filename)
}

func getDownloadTargetBasePath(item *manifestItem, baseDir string) string {
	return fmt.Sprintf("%s/%s", baseDir, item.Path)
}

func must[T any](r T, e error) T {
	if e != nil {
		logger.Fatal(e)
	}

	return r
}
