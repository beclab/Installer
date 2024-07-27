package storage

import (
	"fmt"
	"path"

	"bytetrade.io/web3os/installer/pkg/common"
	corecommon "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/utils"
)

// ~ CheckEtcdSSL
type CheckEtcdSSL struct {
	common.KubePrepare
}

func (p *CheckEtcdSSL) PreCheck(runtime connector.Runtime) (bool, error) {
	var files = []string{
		"/etc/ssl/etcd/ssl/ca.pem",
		fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s-key.pem", runtime.RemoteHost().GetName()),
		fmt.Sprintf("/etc/ssl/etcd/ssl/node-%s.pem", runtime.RemoteHost().GetName()),
	}
	for _, f := range files {
		if !utils.IsExist(f) {
			return false, nil
		}
	}
	return true, nil
}

// ~ CheckStorageType
type CheckStorageType struct {
	common.KubePrepare
	StorageType string
}

func (p *CheckStorageType) PreCheck(runtime connector.Runtime) (bool, error) {
	storageType, _ := p.PipelineCache.GetMustString(common.CacheStorageType)
	if storageType == "" || storageType != p.StorageType {
		return false, nil
	}
	return true, nil
}

// ~ CheckStorageVendor
type CheckStorageVendor struct {
	common.KubePrepare
}

func (p *CheckStorageVendor) PreCheck(runtime connector.Runtime) (bool, error) {
	storageVendor, _ := p.PipelineCache.GetMustString(common.CacheStorageVendor)
	if storageVendor != "true" {
		return false, nil
	}

	if storageType, _ := p.PipelineCache.GetMustString(common.CacheStorageType); storageType != "s3" && storageType != "oss" {
		return false, nil
	}

	if storageBucket, _ := p.PipelineCache.GetMustString(common.CacheStorageBucket); storageBucket == "" {
		return false, nil
	}

	return true, nil
}

// ~ CreateJuiceFsDataPath
type CreateJuiceFsDataPath struct {
	common.KubePrepare
}

func (p *CreateJuiceFsDataPath) PreCheck(runtime connector.Runtime) (bool, error) {
	var juiceFsDataPath = path.Join(corecommon.TerminusDir, "data", "juicefs")
	if !utils.IsExist(juiceFsDataPath) {
		utils.Mkdir(juiceFsDataPath)
	}

	var juiceFsMountPoint = path.Join(corecommon.TerminusDir, "rootfs")
	if !utils.IsExist(juiceFsMountPoint) {
		utils.Mkdir(juiceFsMountPoint)
	}

	var juiceFsCacheDir = path.Join(corecommon.TerminusDir, "jfscache")
	if !utils.IsExist(juiceFsCacheDir) {
		utils.Mkdir(juiceFsCacheDir)
	}

	return true, nil
}
