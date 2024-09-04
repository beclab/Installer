package storage

import (
	"fmt"
	"io/fs"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"
	"bytetrade.io/web3os/installer/version"
	"github.com/pkg/errors"
)

type DownloadStorageBinaries struct {
	common.KubeAction
}

func (t *DownloadStorageBinaries) Execute(runtime connector.Runtime) error {
	var arch = constants.OsArch

	var prePath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	terminus := files.NewKubeBinary("terminus-cli", arch, version.VERSION, path.Join(prePath, cc.WizardDir))
	minio := files.NewKubeBinary("minio", arch, kubekeyapiv1alpha2.DefaultMinioVersion, prePath)
	miniooperator := files.NewKubeBinary("minio-operator", arch, kubekeyapiv1alpha2.DefaultMinioOperatorVersion, prePath)
	redis := files.NewKubeBinary("redis", arch, kubekeyapiv1alpha2.DefaultRedisVersion, prePath)
	juicefs := files.NewKubeBinary("juicefs", arch, kubekeyapiv1alpha2.DefaultJuiceFsVersion, prePath)
	velero := files.NewKubeBinary("velero", arch, kubekeyapiv1alpha2.DefaultVeleroVersion, prePath)

	// gpu
	keyring := files.NewKubeBinary("cuda-keyring", arch, "1.0", prePath)
	gpgkey := files.NewKubeBinary("gpgkey", arch, "", prePath)
	libnvidia := files.NewKubeBinary("libnvidia-container", arch, "", prePath)

	binaries := []*files.KubeBinary{terminus, minio, miniooperator, redis, juicefs, velero, keyring, gpgkey}
	if constants.OsPlatform == common.Ubuntu && !strings.Contains(constants.OsVersion, "24.") {
		binaries = append(binaries, libnvidia)
	}
	// libnvidia
	for _, binary := range binaries {
		if err := binary.CreateBaseDir(); err != nil {
			return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
		}

		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)

		var exists = util.IsExist(binary.Path())
		if exists {
			p := binary.Path()
			if err := binary.SHA256Check(); err != nil {
				_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
			} else {
				logger.Debugf("%s exists", binary.FileName)
			}
			continue
		}

		if !exists || binary.OverWrite {
			logger.Infof("%s downloading %s %s %s ...", common.LocalHost, arch, binary.ID, binary.Version)
			if err := binary.Download(); err != nil {
				return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
			}
		}

		t.PipelineCache.Set(common.KubeBinaries+"-"+arch+"-"+binary.ID, binary)
	}

	return nil
}

type MkStorageDir struct {
	common.KubeAction
}

func (t *MkStorageDir) Execute(runtime connector.Runtime) error {
	if utils.IsExist(StorageDataDir) {
		if utils.IsExist(cc.TerminusDir) {
			_, _ = runtime.GetRunner().SudoCmdExt(fmt.Sprintf("rm -rf %s", cc.TerminusDir), false, false)
		}

		if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("mkdir -p %s", StorageDataTerminusDir), false, false); err != nil {
			return err
		}
		if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("ln -s %s %s", StorageDataTerminusDir, cc.TerminusDir), false, false); err != nil {
			return err
		}
	}

	return nil
}

type DownloadStorageCli struct {
	common.KubeAction
}

func (t *DownloadStorageCli) Execute(runtime connector.Runtime) error {
	var storageType = t.KubeConf.Arg.Storage.StorageType
	var arch = fmt.Sprintf("%s-%s", constants.OsType, constants.OsArch)

	var prePath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	var binary *files.KubeBinary
	switch storageType {
	case "s3":
		binary = files.NewKubeBinary("awscli", arch, "", prePath)
	case "oss":
		binary = files.NewKubeBinary("ossutil", arch, kubekeyapiv1alpha2.DefaultOssUtilVersion, prePath)
	default:
		return nil
	}

	binaries := []*files.KubeBinary{binary}
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

	t.PipelineCache.Set(common.KubeBinaries+"-"+arch, binariesMap)

	return nil
}

type UnMountS3 struct {
	common.KubeAction
}

func (t *UnMountS3) Execute(runtime connector.Runtime) error {
	// exp https://terminus-os-us-west-1.s3.us-west-1.amazonaws.com
	// s3  s3://terminus-os-us-west-1

	storageBucket := t.KubeConf.Arg.Storage.StorageBucket
	storageAccessKey := t.KubeConf.Arg.Storage.StorageAccessKey
	storageSecretKey := t.KubeConf.Arg.Storage.StorageSecretKey
	storageToken := t.KubeConf.Arg.Storage.StorageToken
	storageClusterId := t.KubeConf.Arg.Storage.StorageClusterId

	_, a, f := strings.Cut(storageBucket, "://")
	if !f {
		logger.Errorf("get s3 bucket failed %s", storageBucket)
		return nil
	}
	sa := strings.Split(a, ".")
	if len(sa) < 2 {
		logger.Errorf("get s3 bucket failed %s", storageBucket)
		return nil
	}
	endpoint := fmt.Sprintf("s3://%s", sa[0])
	var cmd = fmt.Sprintf("AWS_ACCESS_KEY_ID=%s AWS_SECRET_ACCESS_KEY=%s AWS_SESSION_TOKEN=%s /usr/local/bin/aws s3 rm %s/%s --recursive",
		storageAccessKey, storageSecretKey, storageToken, endpoint, storageClusterId,
	)

	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		logger.Errorf("failed to unmount s3 bucket %s: %v", storageBucket, err)
		return err
	}

	return nil
}

type UnMountOSS struct {
	common.KubeAction
}

func (t *UnMountOSS) Execute(runtime connector.Runtime) error {
	storageBucket := t.KubeConf.Arg.Storage.StorageBucket
	storageAccessKey := t.KubeConf.Arg.Storage.StorageAccessKey
	storageSecretKey := t.KubeConf.Arg.Storage.StorageSecretKey
	storageToken := t.KubeConf.Arg.Storage.StorageToken
	storageClusterId := t.KubeConf.Arg.Storage.StorageClusterId

	// exp: https://name.area.aliyuncs.com
	// oss  oss://name
	// endpoint: https://area.aliyuncs.com

	b, a, f := strings.Cut(storageBucket, "://")
	if !f {
		logger.Errorf("get oss bucket failed %s", storageBucket)
		return nil
	}

	s := strings.Split(a, ".")
	if len(s) != 4 {
		logger.Errorf("get oss bucket failed %s", storageBucket)
		return nil
	}
	ossName := fmt.Sprintf("oss://%s", s[0])
	ossEndpoint := fmt.Sprintf("%s://%s.%s.%s", b, s[1], s[2], s[3])

	var cmd = fmt.Sprintf("/usr/local/sbin/ossutil64 rm %s/%s/ --endpoint=%s --access-key-id=%s --access-key-secret=%s --sts-token=%s -r -f", ossName, storageClusterId, ossEndpoint, storageAccessKey, storageSecretKey, storageToken)

	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, true); err != nil {
		logger.Errorf("failed to unmount oss bucket %s: %v", storageBucket, err)
	}

	return nil
}

type StopJuiceFS struct {
	common.KubeAction
}

func (t *StopJuiceFS) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmdExt("systemctl stop juicefs; systemctl disable juicefs", false, false)

	_, _ = runtime.GetRunner().SudoCmdExt("rm -rf /var/jfsCache /terminus/jfscache", false, false)

	_, _ = runtime.GetRunner().SudoCmdExt("umount /terminus/rootfs", false, false)

	_, _ = runtime.GetRunner().SudoCmdExt("rm -rf /terminus/rootfs", false, false)

	return nil
}

type StopMinio struct {
	common.KubeAction
}

func (t *StopMinio) Execute(runtime connector.Runtime) error {
	_, _ = runtime.GetRunner().SudoCmdExt("systemctl stop minio; systemctl disable minio", false, false)
	return nil
}

type StopMinioOperator struct {
	common.KubeAction
}

func (t *StopMinioOperator) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl stop minio-operator; systemctl disable minio-operator"
	_, _ = runtime.GetRunner().SudoCmdExt(cmd, false, false)
	return nil
}

type StopRedis struct {
	common.KubeAction
}

func (t *StopRedis) Execute(runtime connector.Runtime) error {
	var cmd = "systemctl stop redis-server; systemctl disable redis-server"
	_, _ = runtime.GetRunner().SudoCmdExt(cmd, false, false)
	_, _ = runtime.GetRunner().SudoCmdExt("killall -9 redis-server", false, false)
	_, _ = runtime.GetRunner().SudoCmdExt("unlink /usr/bin/redis-server; unlink /usr/bin/redis-cli", false, false)

	return nil
}

type RemoveTerminusFiles struct {
	common.KubeAction
}

func (t *RemoveTerminusFiles) Execute(runtime connector.Runtime) error {
	var files = []string{
		"/usr/local/bin/redis-*",
		"/usr/bin/redis-*",
		"/sbin/mount.juicefs",
		"/etc/init.d/redis-server",
		"/usr/local/bin/juicefs",
		"/usr/local/bin/minio",
		"/usr/local/bin/velero",
		"/etc/systemd/system/redis-server.service",
		"/etc/systemd/system/minio.service",
		"/etc/systemd/system/minio-operator.service",
		"/etc/systemd/system/juicefs.service",
		"/terminus/",
	}

	for _, f := range files {
		runtime.GetRunner().SudoCmdExt(fmt.Sprintf("rm -rf %s", f), false, true)
	}

	return nil
}

type DeleteTmp struct {
	common.KubeAction
}

func (t *DeleteTmp) Execute(runtime connector.Runtime) error {
	var tmpPath = path.Join(common.RootDir, "tmp", "install_log")
	if util.IsExist(tmpPath) {
		util.RemoveDir(tmpPath)
	}
	return nil
}

type DeleteCaches struct {
	common.KubeAction
	BaseDir string
	Skip    bool
}

func (t *DeleteCaches) Execute(runtime connector.Runtime) error {
	if t.Skip {
		return nil
	}
	home := runtime.GetHomeDir()
	baseDir := t.BaseDir
	if baseDir == "" {
		baseDir = home + "/.terminus"
	}

	var cachesDirs []string

	filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if path != baseDir {
			if d.IsDir() {
				cachesDirs = append(cachesDirs, path)
				return filepath.SkipDir
			}
		}
		return nil
	},
	)

	if cachesDirs != nil && len(cachesDirs) > 0 {
		for _, cachesDir := range cachesDirs {
			if util.IsExist(cachesDir) {
				util.RemoveDir(cachesDir)
			}
		}
	}

	return nil
}
