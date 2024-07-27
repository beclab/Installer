package storage

import (
	"path"

	"bytetrade.io/web3os/installer/pkg/core/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
)

var (
	Root                   = path.Join("/")
	StorageDataDir         = path.Join(Root, "osdata")
	StorageDataTerminusDir = path.Join(StorageDataDir, common.TerminusDir)

	RedisRootDir             = path.Join(Root, cc.TerminusDir, "data", "redis")
	RedisConfigDir           = path.Join(RedisRootDir, "etc")
	RedisDataDir             = path.Join(RedisRootDir, "data")
	RedisLogDir              = path.Join(RedisRootDir, "log")
	RedisRunDir              = path.Join(RedisRootDir, "run")
	RedisConfigFile          = path.Join(RedisConfigDir, "redis.conf")
	RedisServiceFile         = path.Join(Root, "etc", "systemd", "system", "redis-server.service")
	RedisServerFile          = path.Join(Root, "usr", "bin", "redis-server")
	RedisCliFile             = path.Join(Root, "usr", "bin", "redis-cli")
	RedisServerInstalledFile = path.Join(Root, "usr", "local", "bin", "redis-server")
	RedisCliInstalledFile    = path.Join(Root, "usr", "local", "bin", "redis-cli")

	JuiceFsFile          = path.Join(Root, "usr", "local", "bin", "juicefs")
	JuiceFsDataDir       = path.Join(Root, cc.TerminusDir, "data", "juicefs")
	JuiceFsCacheDir      = path.Join(Root, cc.TerminusDir, "jfscache")
	JuiceFsMountPointDir = path.Join(Root, cc.TerminusDir, "rootfs")
	JuiceFsServiceFile   = path.Join(Root, "etc", "systemd", "system", "juicefs.service")

	MinioRootUser    = "minioadmin"
	MinioDataDir     = path.Join(Root, cc.TerminusDir, "data", "minio", "vol1")
	MinioFile        = path.Join(Root, "usr", "local", "bin", "minio")
	MinioServiceFile = path.Join(Root, "etc", "systemd", "system", "minio.service")
	MinioConfigFile  = path.Join(Root, "etc", "default", "minio")

	MinioOperatorFile = path.Join(Root, "usr", "local", "bin", "minio-operator")
)
