package storage

import (
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"context"
	"fmt"
	"strings"
	"time"

	"bytetrade.io/web3os/installer/pkg/common"
	cc "bytetrade.io/web3os/installer/pkg/core/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/manifest"
	"bytetrade.io/web3os/installer/pkg/utils"

	redisTemplates "bytetrade.io/web3os/installer/pkg/storage/templates"
	"github.com/pkg/errors"
)

type CheckRedisServiceState struct {
	common.KubeAction
}

func (t *CheckRedisServiceState) Execute(runtime connector.Runtime) error {
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	var rpwd, _ = t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	var cmd = fmt.Sprintf("%s -h %s -a %s ping", RedisCliFile, localIp, rpwd)
	if pong, _ := runtime.GetRunner().Host.SudoCmd(cmd, false, false); !strings.Contains(pong, "PONG") {
		return fmt.Errorf("failed to connect redis server: %s:6379", localIp)
	}

	return nil
}

type EnableRedisService struct {
	common.KubeAction
}

func (t *EnableRedisService) Execute(runtime connector.Runtime) error {
	if _, err := runtime.GetRunner().Host.SudoCmd("sysctl -w vm.overcommit_memory=1 net.core.somaxconn=10240", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl daemon-reload", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl restart redis-server", false, false); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd("systemctl enable redis-server", false, false); err != nil {
		return err
	}

	var cmd = "( sleep 10 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server ) || ( systemctl restart redis-server && sleep 3 && systemctl --no-pager status redis-server )"
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return err
	}

	return nil
}

type GetOrSetRedisPassword struct {
	common.KubeAction
}

func (t *GetOrSetRedisPassword) Execute(runtime connector.Runtime) (err error) {
	var redisPassword string
	defer func() {
		if err == nil {
			t.PipelineCache.Set(common.CacheHostRedisPassword, redisPassword)
		}
	}()
	if !util.IsExist(RedisConfigFile) {
		redisPassword, _ = utils.GeneratePassword(16)
		return
	}
	redisPassword, err = getRedisPwdFromConfigFile()
	if err != nil {
		return
	}
	if redisPassword != "" {
		logger.Debugf("using existing Redis password: %s found in %s", redisPassword, RedisConfigFile)
		return
	}
	logger.Warnf("found Redis config file %s but password is not set, generating a new one", RedisConfigFile)
	redisPassword, _ = utils.GeneratePassword(16)
	return
}

func getRedisPwdFromConfigFile() (string, error) {
	var cmd = fmt.Sprintf("cat %s 2>&1 |grep requirepass |cut -d' ' -f2 |tr -d '\n'", RedisConfigFile)
	if res, _, err := util.Exec(context.Background(), cmd, false, false); err != nil {
		return "", errors.Wrap(err, "failed to get redis password")
	} else {
		return res, nil
	}
}

type ConfigRedis struct {
	common.KubeAction
}

func (t *ConfigRedis) Execute(runtime connector.Runtime) error {
	redisPassword, ok := t.PipelineCache.GetMustString(common.CacheHostRedisPassword)
	if !ok || redisPassword == "" {
		return errors.New("redis password not set")
	}
	var systemInfo = runtime.GetSystemInfo()
	var localIp = systemInfo.GetLocalIp()
	if !utils.IsExist(RedisRootDir) {
		utils.Mkdir(RedisConfigDir)
		utils.Mkdir(RedisDataDir)
		utils.Mkdir(RedisLogDir)
		utils.Mkdir(RedisRunDir)
	}

	var data = util.Data{
		"LocalIP":  localIp,
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

	return nil
}

type CheckRedisExists struct {
	common.KubePrepare
}

func (p *CheckRedisExists) PreCheck(runtime connector.Runtime) (bool, error) {
	if !utils.IsExist(RedisServerInstalledFile) {
		return true, nil
	}

	if !utils.IsExist(RedisServerFile) {
		return true, nil
	}

	return false, nil
}

type InstallRedis struct {
	common.KubeAction
	manifest.ManifestAction
}

func (t *InstallRedis) Execute(runtime connector.Runtime) error {
	redis, err := t.Manifest.Get("redis")
	if err != nil {
		return err
	}

	path := redis.FilePath(t.BaseDir)

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("rm -rf /tmp/redis-* && cp -f %s /tmp/%s && cd /tmp && tar xf ./%s", path, redis.Filename, redis.Filename), false, false); err != nil {
		return errors.Wrapf(errors.WithStack(err), "untar redis failed")
	}

	unpackPath := strings.TrimSuffix(redis.Filename, ".tar.gz")
	var cmd = fmt.Sprintf("cd /tmp/%s && cp ./* /usr/local/bin/ && rm -rf ./%s",
		unpackPath, unpackPath)
	if _, err := runtime.GetRunner().Host.SudoCmd(cmd, false, false); err != nil {
		return err
	}
	// if _, err := runtime.GetRunner().Host.SudoCmd("[[ ! -f /usr/local/bin/redis-sentinel ]] && /usr/local/bin/redis-server /usr/local/bin/redis-sentinel || true", false, true); err != nil {
	// 	return err
	// }
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("[[ ! -f %s ]] && ln -s %s %s || true", RedisServerFile, RedisServerInstalledFile, RedisServerFile), false, true); err != nil {
		return err
	}
	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("[[ ! -f %s ]] && ln -s %s %s || true", RedisCliFile, RedisCliInstalledFile, RedisCliFile), false, true); err != nil {
		return err
	}

	return nil
}

type InstallRedisModule struct {
	common.KubeModule
	manifest.ManifestModule
	Skip bool
}

func (m *InstallRedisModule) IsSkip() bool {
	return m.Skip
}

func (m *InstallRedisModule) Init() {
	m.Name = "InstallRedis"

	installRedis := &task.RemoteTask{
		Name:    "Install",
		Hosts:   m.Runtime.GetAllHosts(),
		Prepare: &CheckRedisExists{},
		Action: &InstallRedis{
			ManifestAction: manifest.ManifestAction{
				BaseDir:  m.BaseDir,
				Manifest: m.Manifest,
			},
		},
		Parallel: false,
		Retry:    0,
	}

	getOrSetRedisPassword := &task.LocalTask{
		Name:   "GetOrSetRedisPassword",
		Action: new(GetOrSetRedisPassword),
		Retry:  1,
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
		installRedis,
		getOrSetRedisPassword,
		configRedis,
		enableRedisService,
		checkRedisServiceState,
	}
}
