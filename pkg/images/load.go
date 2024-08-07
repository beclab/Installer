package images

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/cavaliergopher/grab/v3"
)

const MAX_IMPORT_RETRY int = 5

type CheckImageManifest struct {
	common.KubePrepare
}

func (p *CheckImageManifest) PreCheck(runtime connector.Runtime) (bool, error) {
	var imageManifest = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir, cc.ManifestImage)

	if utils.IsExist(imageManifest) {
		return true, nil
	}
	return false, fmt.Errorf("image manifest not exist")
}

type LoadImages struct {
	common.KubeAction
}

func (t *LoadImages) Execute(runtime connector.Runtime) (reserr error) {
	var kubeConf = t.KubeConf
	var host = runtime.RemoteHost()
	var imageManifest = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ManifestDir, cc.ManifestImage)
	mf, _ := readImageManifest(imageManifest)
	if mf == nil {
		logger.Debugf("image manifest is empty, skip load images")
		return nil
	}

	var imagesDir = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.ImageCacheDir)
	if err := utils.Mkdir(imagesDir); err != nil {
		return fmt.Errorf("create images cache directory error %v", err)
	}

	retry := func(f func() error, times int) (err error) {
		for i := 0; i < times; i++ {
			err = f()
			if err == nil {
				return nil
			}

			var dur = 5 + (i+1)*10
			// fmt.Printf("import %s failed, wait for %d seconds(%d times)\n", err, dur, i+1)
			logger.Errorf("import error %v, wait for %d seconds(%d times)", err, dur, i+1)
			if (i + 1) < times {
				time.Sleep(time.Duration(dur) * time.Second)
			}
		}
		return
	}

	// var i = 0
	for _, imageRepoTag := range mf {
		// i++
		// if i > 1 {
		// 	break
		// }
		reserr = nil
		if inspectImage(runtime.GetRunner(), kubeConf.Cluster.Kubernetes.ContainerManager, imageRepoTag) == nil {
			logger.Debugf("%s already exists", imageRepoTag)
			continue
		}

		// logger.Debugf("preloading %s", imageRepoTag)
		var start = time.Now()
		var imageHashTag = utils.MD5(imageRepoTag)
		var imageFileName string

		var found = false
		filepath.Walk(imagesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if !strings.HasPrefix(info.Name(), imageHashTag) ||
				!HasSuffixI(info.Name(), ".tar.gz", ".tgz", ".tar") {
				return nil
			}

			if strings.HasPrefix(info.Name(), imageHashTag) {
				found = true
				imageFileName = path
				return filepath.SkipDir
			}

			return nil
		})

		if !found {
			imageFileName = path.Join(imagesDir, fmt.Sprintf("%s.tar.gz", imageHashTag))
			if err := downloadImageFile(host.GetArch(), imageRepoTag, imageFileName); err != nil {
				// logger.Errorf("download image %s(hash:%s) file error %v", imageRepoTag, imageHashTag, err)
				reserr = fmt.Errorf("download image %s(hash:%s) file error %v", imageRepoTag, imageHashTag, err)
				break
			} else {
				logger.Debugf("download %s success", imageRepoTag)
			}
		}

		var imgFileName = filepath.Base(imageFileName)
		var loadCmd string
		switch kubeConf.Cluster.Kubernetes.ContainerManager {
		case "crio":
			loadCmd = "ctr" // not implement
		case "containerd":
			if HasSuffixI(imgFileName, ".tar.gz", ".tgz") {
				loadCmd = "env PATH=$PATH gunzip -c %s | ctr -n k8s.io images import -"
			} else {
				loadCmd = "env PATH=$PATH ctr -n k8s.io images import %s"
			}
		case "isula":
			loadCmd = "isula" // not implement
		default:
			if HasSuffixI(imgFileName, ".tar.gz", ".tgz") {
				loadCmd = "docker load"
			} else {
				loadCmd = "docker load -i"
			}
		}

		if err := retry(func() error {
			if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf(loadCmd, imageFileName), false, false); err != nil {
				return fmt.Errorf("%s(%s) error: %v", imageRepoTag, imgFileName, err)
			} else {
				logger.Debugf("import %s success (%s)", imageRepoTag, time.Since(start))
			}
			return nil
		}, MAX_IMPORT_RETRY); err != nil {
			reserr = fmt.Errorf("%s(%s) error: %v", imageRepoTag, imgFileName, err)
			break
		}
	}
	return
}

func inspectImage(runner *connector.Runner, containerManager, imageRepoTag string) error {
	var inspectCmd string = "docker"
	switch containerManager {
	case "crio": //  not implement
		inspectCmd = "ctr"
	case "containerd":
		inspectCmd = "crictl inspecti -q %s"
	case "isula": // not implement
		inspectCmd = "isula"
	default:
		inspectCmd = "docker image inspect %s"
	}

	var cmd = fmt.Sprintf(inspectCmd, imageRepoTag)
	if _, err := runner.SudoCmdExt(cmd, false, false); err != nil {
		return fmt.Errorf("inspect %s error %v", imageRepoTag, err)
	}

	return nil
}

func downloadImageFile(arch, imageRepoTag, imageFilePath string) error {
	var err error
	if arch == common.Amd64 {
		arch = ""
	} else {
		arch = arch + "/"
	}

	var imageFileName = path.Base(imageFilePath)
	var url = fmt.Sprintf("https://dc3p1870nn3cj.cloudfront.net/%s%s", arch, imageFileName)
	for i := 5; i > 0; i-- {
		totalSize, _ := getImageFileSize(url)
		if totalSize > 0 {
			logger.Infof("get image %s size: %s", imageRepoTag, utils.FormatBytes(totalSize))
		}

		client := grab.NewClient()
		req, _ := grab.NewRequest(imageFilePath, url)
		// req.RateLimiter = NewLimiter(1024 * 1024)
		req.HTTPRequest = req.HTTPRequest.WithContext(context.Background())
		ctx, cancel := context.WithTimeout(req.HTTPRequest.Context(), 5*time.Minute)
		defer cancel()

		req.HTTPRequest = req.HTTPRequest.WithContext(ctx)
		resp := client.Do(req)

		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()

		var downloaded int64
	Loop:
		for {
			select {
			case <-t.C:
				downloaded = resp.BytesComplete()
				var progressInfo string
				if totalSize != 0 {
					result := float64(downloaded) / float64(totalSize)
					progressInfo = fmt.Sprintf("transferred %s %s / %s (%.2f%%) / speed: %s", imageFileName, utils.FormatBytes(resp.BytesComplete()), utils.FormatBytes(totalSize), math.Round(result*10000)/100, utils.FormatBytes(int64(resp.BytesPerSecond())))
					logger.Info(progressInfo)
				} else {
					progressInfo = fmt.Sprintf("transferred %s %s / speed: %s\n", imageFileName, utils.FormatBytes(resp.BytesComplete()), utils.FormatBytes(int64(resp.BytesPerSecond())))
					logger.Infof(progressInfo)
				}
			case <-resp.Done:
				break Loop
			}
		}

		if err = resp.Err(); err != nil {
			logger.Infof("download %s error %v", imageFileName, err)
			time.Sleep(2 * time.Second)
			continue
		}
	}

	return err
}

func pullImage(runner *connector.Runner, containerManager, imageRepoTag, imageHashTag, dst string) error {
	var pullCmd string = "docker"
	var inspectCmd string = "docker"
	var exportCmd string = "docker"
	switch containerManager {
	case "crio": // not implement
		pullCmd = "ctr"
		inspectCmd = "ctr"
		exportCmd = "ctr"
	case "containerd":
		pullCmd = "crictl pull %s"
		inspectCmd = "crictl inspecti -q %s"
		exportCmd = "ctr -n k8s.io image export %s %s"
	case "isula": // not implement
		pullCmd = "isula"
		inspectCmd = "isula"
		exportCmd = "isula"
	default:
		pullCmd = "docker pull %s"
		exportCmd = "docker save -o %s %s"
	}

	var cmd = fmt.Sprintf(pullCmd, imageRepoTag)
	if _, err := runner.SudoCmdExt(cmd, false, false); err != nil {
		return fmt.Errorf("pull %s error %v", imageRepoTag, err)
	}

	var repoTag = imageRepoTag
	if containerManager == "containerd" {
		cmd = fmt.Sprintf(inspectCmd, imageRepoTag)
		stdout, err := runner.SudoCmdExt(cmd, false, false)
		if err != nil {
			return fmt.Errorf("inspect %s error %v", imageRepoTag, err)
		}
		var ii ImageInspect
		if err = json.Unmarshal([]byte(stdout), &ii); err != nil {
			return fmt.Errorf("unmarshal %s error %v", imageRepoTag, err)
		}
		repoTag = ii.Status.RepoTags[0]
	}

	var dstFile = path.Join(dst, fmt.Sprintf("%s.tar", imageHashTag))
	cmd = fmt.Sprintf(exportCmd, dstFile, repoTag)
	if _, err := runner.SudoCmdExt(cmd, false, false); err != nil {
		return fmt.Errorf("export %s error: %v", imageRepoTag, err)
	}
	if _, err := runner.SudoCmdExt(fmt.Sprintf("gzip %s", dstFile), false, false); err != nil {
		return fmt.Errorf("gzip %s error: %v", dstFile, err)
	}

	return nil
}

func getImageFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("bad status: %s", resp.Status)
	}

	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return -1, fmt.Errorf("failed to parse content length: %v, header: %s", err, resp.Header.Get("Content-Length"))
	}
	return size, nil
}

type RateLimiter struct {
	r, n int
}

func NewLimiter(r int) grab.RateLimiter {
	return &RateLimiter{r: r}
}

func (c *RateLimiter) WaitN(ctx context.Context, n int) (err error) {
	c.n += n
	time.Sleep(
		time.Duration(1.00 / float64(c.r) * float64(n) * float64(time.Second)))
	return
}
