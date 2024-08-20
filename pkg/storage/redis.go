package storage

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/constants"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/files"
	"bytetrade.io/web3os/installer/pkg/utils"

	redisTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"github.com/pkg/errors"
)

type CheckRedisServiceState struct {
	common.KubeAction
}

func (t *CheckRedisServiceState) Execute(runtime connector.Runtime) error {
	var rpwd, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var cmd = fmt.Sprintf("%s -h %s -a %s ping", RedisCliFile, constants.LocalIp, rpwd)
	if pong, _ := runtime.GetRunner().SudoCmd(cmd, false, false); !strings.Contains(pong, "PONG") {
		return fmt.Errorf("failed to connect redis server: %s:6379", constants.LocalIp)
	}

	return nil
}

type EnableRedisService struct {
	common.KubeAction
}

func (t *EnableRedisService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().SudoCmdExt("sysctl -w vm.overcommit_memory=1 net.core.somaxconn=10240", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmdExt("systemctl daemon-reload", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmdExt("systemctl restart redis-server", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmdExt("systemctl enable redis-server", false, false); err != nil {
		return err
	}

	var cmd = "( sleep 10 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server )"
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, false); err != nil {
		return err
	}

	cmd = fmt.Sprintf("awk '/requirepass/{print \\$NF}' %s", RedisConfigFile)
	rpwd, _ := runtime.GetRunner().SudoCmd(cmd, false, false)
	if rpwd == "" {
		return fmt.Errorf("get redis password failed")
	}

	t.PipelineCache.Set(common.CacheHostRedisPassword, rpwd)

	return nil
}

type ConfigRedis struct {
	common.KubeAction
}

func (t *ConfigRedis) Execute(runtime connector.Runtime) error {
	var redisPassword, _ = utils.GeneratePassword(16) // todo
	if !utils.IsExist(RedisRootDir) {
		utils.Mkdir(RedisConfigDir)
		utils.Mkdir(RedisDataDir)
		utils.Mkdir(RedisLogDir)
		utils.Mkdir(RedisRunDir)
	}

	var data = util.Data{
		"LocalIP":  constants.LocalIp,
		"RootPath": RedisRootDir,
		"Password": redisPassword,
	}
	redisConfStr, err := util.Render(redisTemplates.RedisConf, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render redis conf template failed")
	}
	if err := util.WriteFile(RedisConfigFile, []byte(redisConfStr), cc.FileMode0640); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write redis conf %s failed", RedisConfigFile))
	}

	data = util.Data{
		"RedisBinPath":  RedisServerFile,
		"RootPath":      RedisRootDir,
		"RedisConfPath": RedisConfigFile,
	}
	redisServiceStr, err := util.Render(redisTemplates.RedisService, data)
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "render redis service template failed")
	}
	if err := util.WriteFile(RedisServiceFile, []byte(redisServiceStr), cc.FileMode0644); err != nil {
		return errors.Wrap(errors.WithStack(err), fmt.Sprintf("write redis service %s failed", RedisServiceFile))
	}

	t.PipelineCache.Set(common.CacheHostRedisPassword, redisPassword)

	return nil
}

type DownloadRedis struct {
	common.KubeAction
}

func (t *DownloadRedis) Execute(runtime connector.Runtime) error {
	var arch = constants.OsArch
	var prePath = path.Join(runtime.GetHomeDir(), cc.TerminusKey, cc.PackageCacheDir)
	binary := files.NewKubeBinary("redis", arch, kubekeyapiv1alpha2.DefaultRedisVersion, prePath)

	if err := binary.CreateBaseDir(); err != nil {
		return errors.Wrapf(errors.WithStack(err), "create file %s base dir failed", binary.FileName)
	}

	var exists = util.IsExist(binary.Path())
	if exists {
		p := binary.Path()
		if err := binary.SHA256Check(); err != nil {
			_ = exec.Command("/bin/sh", "-c", fmt.Sprintf("rm -f %s", p)).Run()
		}
	}

	if !exists || binary.OverWrite {
		logger.Infof("%s downloading %s %s %s ...", common.LocalHost, runtime.RemoteHost().GetArch(), binary.ID, binary.Version)
		if err := binary.Download(); err != nil {
			return fmt.Errorf("Failed to download %s binary: %s error: %w ", binary.ID, binary.Url, err)
		}
	}

	t.PipelineCache.Set(common.KubeBinaries+"-"+arch+"-"+"redis", binary)

	return nil
}

type InstallRedis struct {
	common.KubeAction
}

func (t *InstallRedis) Execute(runtime connector.Runtime) error {
	var arch = constants.OsArch
	redisObj, ok := t.PipelineCache.Get(common.KubeBinaries + "-" + arch + "-" + "redis")
	if !ok {
		return errors.New("get Redis Binary by pipeline cache failed")
	}

	redis := redisObj.(*files.KubeBinary)

	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("cd %s && tar xf ./%s", redis.BaseDir, redis.FileName), false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "untar redis failed")
	}

	var cmd = fmt.Sprintf("cd %s/redis-%s && cp ./* /usr/local/bin/ && ln -s /usr/local/bin/redis-server /usr/local/bin/redis-sentinel && rm -rf ./redis-%s", redis.BaseDir, redis.Version, redis.Version)
	if _, err := runtime.GetRunner().SudoCmdExt(cmd, false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("ln -s %s %s", RedisServerInstalledFile, RedisServerFile), false, true); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().SudoCmdExt(fmt.Sprintf("ln -s %s %s", RedisCliInstalledFile, RedisCliFile), false, true); err != nil {
		return err
	}

	return nil
}

type InstallRedisModule struct {
	common.KubeModule
}

func (m *InstallRedisModule) Init() {
	m.Name = "InstallRedis"

	downloadRedis := &task.RemoteTask{
		Name:     "Download",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(DownloadRedis),
		Parallel: false,
		Retry:    1,
	}

	installRedis := &task.RemoteTask{
		Name:     "Install",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(InstallRedis),
		Parallel: false,
		Retry:    0,
	}

	configRedis := &task.RemoteTask{
		Name:     "Config",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(ConfigRedis),
		Parallel: false,
		Retry:    0,
	}

	enableRedisService := &task.RemoteTask{
		Name:     "EnableRedisService",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(EnableRedisService),
		Parallel: false,
		Retry:    0,
	}

	checkRedisServiceState := &task.RemoteTask{
		Name:     "CheckState",
		Hosts:    m.Runtime.GetAllHosts(),
		Action:   new(CheckRedisServiceState),
		Parallel: false,
		Retry:    3,
		Delay:    3 * time.Second,
	}

	m.Tasks = []task.Interface{
		downloadRedis,
		installRedis,
		configRedis,
		enableRedisService,
		checkRedisServiceState,
	}
}
