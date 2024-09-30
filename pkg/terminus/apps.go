package terminus

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"bytetrade.io/web3os/installer/pkg/clientset"
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/task"
	"bytetrade.io/web3os/installer/pkg/core/util"
	"bytetrade.io/web3os/installer/pkg/utils"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var BuiltInApps = []string{"portfolio", "vault", "desktop", "message", "wise", "search", "appstore", "notification", "dashboard", "settings", "profile", "agent", "files"}

type InstallAppsModule struct {
	common.KubeModule
}

func (i *InstallAppsModule) Init() {
	i.Name = "Install Built-in apps"
	i.Desc = "Install Built-in apps"

	prepareAppValues := &task.LocalTask{
		Name:   "PrepareAppValues",
		Desc:   "Prepare app values",
		Action: new(PrepareAppValues),
	}

	installApps := &task.LocalTask{
		Name:   "InstallApps",
		Desc:   "Install apps",
		Action: new(InstallApps),
	}

	cleearAppsValues := &task.LocalTask{
		Name:   "CleearAppsValues",
		Desc:   "Cleear apps values",
		Action: new(CleearAppsValues),
	}

	copyFiles := &task.LocalTask{
		Name:   "CopyFiles",
		Desc:   "Copy files",
		Action: new(CopyFiles),
		Retry:  5,
		Delay:  5 * time.Second,
	}

	i.Tasks = []task.Interface{
		prepareAppValues,
		installApps,
		cleearAppsValues,
		copyFiles,
	}

}

type PrepareAppValues struct {
	common.KubeAction
}

func (u *PrepareAppValues) Execute(runtime connector.Runtime) error {
	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ns := fmt.Sprintf("user-space-%s", u.KubeConf.Arg.User.UserName)
	redisPassword, err := getRedisPassword(client, runtime)
	if err != nil {
		return err
	}

	bfDocUrl, _ := getDocUrl(ctx, runtime)
	bflNodeName, err := getBflPod(ctx, ns, client, runtime)
	if err != nil {
		return err
	}
	bflAnnotations, err := getBflAnnotation(ctx, ns, client, runtime)
	if err != nil {
		return err
	}
	fsType := getFsType(u.KubeConf.Arg.WSL)
	gpuType := getGpuType(u.KubeConf.Arg.GPU.Enable, u.KubeConf.Arg.GPU.Share)
	appValues := getAppSecrets(getAppPatches())

	var values = map[string]interface{}{
		"bfl": map[string]interface{}{
			"nodeport":               30883,
			"nodeport_ingress_http":  30083,
			"nodeport_ingress_https": 30082,
			"username":               u.KubeConf.Arg.User.UserName,
			"admin_user":             true,
			"url":                    bfDocUrl,
			"nodeName":               bflNodeName,
		},
		"pvc": map[string]string{
			"userspace": bflAnnotations["userspace_pv"],
		},
		"userspace": map[string]string{
			"userData": fmt.Sprintf("%s/Home", bflAnnotations["userspace_hostpath"]),
			"appData":  fmt.Sprintf("%s/Data", bflAnnotations["userspace_hostpath"]),
			"appCache": bflAnnotations["appcache_hostpath"],
			"dbdata":   bflAnnotations["dbdata_hostpath"],
		},
		"desktop": map[string]interface{}{
			"nodeport": 30180,
		},
		"global": map[string]interface{}{
			"bfl": map[string]interface{}{
				"username": u.KubeConf.Arg.User.UserName,
			},
		},
		"debugVersion": os.Getenv("DEBUG_VERSION"),
		"gpu":          gpuType,
		"fs_type":      fsType,
		"os":           appValues,
		"kubesphere": map[string]string{
			"redis_password": redisPassword,
		},
	}

	u.ModuleCache.Set(common.CacheAppValues, values)

	return nil
}

type InstallApps struct {
	common.KubeAction
}

func (i *InstallApps) Execute(runtime connector.Runtime) error {
	var appPath = path.Join(runtime.GetInstallerDir(), "wizard", "config", "apps")
	var appDirs []string
	filepath.WalkDir(appPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() || d.Name() == appPath {
			return nil
		}

		appDirs = append(appDirs, path)

		return nil
	})

	var ns = fmt.Sprintf("user-space-%s", i.KubeConf.Arg.User.UserName)
	config, err := ctrl.GetConfig()
	if err != nil {
		return err
	}
	actionConfig, settings, err := utils.InitConfig(config, ns)
	if err != nil {
		return err
	}

	var parms = make(map[string]interface{})
	var values, ok = i.ModuleCache.Get(common.CacheAppValues)
	if !ok {
		return fmt.Errorf("app values not found")
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	parms["set"] = values
	parms["force"] = true
	for _, appDir := range appDirs {
		appName := filepath.Base(appDir)
		if err := utils.UpgradeCharts(ctx, actionConfig, settings, appName, appDir, "", ns, parms, false); err != nil {
			logger.Errorf("upgrade %s failed %v", appName, err)
		}
	}

	return nil
}

type CleearAppsValues struct {
	common.KubeAction
}

func (c *CleearAppsValues) Execute(runtime connector.Runtime) error {
	// clear apps values.yaml
	_, _ = runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("cat /dev/null > %s/wizard/config/apps/values.yaml", runtime.GetInstallerDir()), false, false)

	_, _ = runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("cat /dev/null > %s/wizard/config/launcher/values.yaml", runtime.GetInstallerDir()), false, false)

	return nil
}

type CopyFiles struct {
	common.KubeAction
}

func (c *CopyFiles) Execute(runtime connector.Runtime) error {
	kubectl, _ := util.GetCommand(common.CommandKubectl)

	if _, err := runtime.GetRunner().Host.SudoCmd(fmt.Sprintf("%s get --raw='/readyz?verbose'", kubectl), false, false); err != nil {
		return fmt.Errorf("%s is not health yet, please check it", c.KubeConf.Cluster.Kubernetes.Type)
	}

	client, err := clientset.NewKubeClient()
	if err != nil {
		return errors.Wrap(errors.WithStack(err), "kubeclient create error")
	}

	appServiceName, err := getAppServiceName(client, runtime)
	if err != nil {
		return err
	}

	kubeclt, _ := util.GetCommand(common.CommandKubectl)
	for _, app := range []string{"launcher", "apps"} {
		var cmd = fmt.Sprintf("%s cp %s/wizard/config/%s os-system/%s:/userapps -c app-service", kubeclt, runtime.GetInstallerDir(), app, appServiceName)
		if _, err = runtime.GetRunner().Host.SudoCmd(cmd, false, true); err != nil {
			return errors.Wrap(errors.WithStack(err), "copy files failed")
		}
	}

	return nil
}

func getAppServiceName(client clientset.Client, runtime connector.Runtime) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pods, err := client.Kubernetes().CoreV1().Pods(common.NamespaceOsSystem).List(ctx, metav1.ListOptions{LabelSelector: "tier=app-service"})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get app-service failed")
	}

	if len(pods.Items) == 0 {
		return "", errors.New("app-service not found")
	}

	return pods.Items[0].Name, nil
}

func getBflPod(ctx context.Context, ns string, client clientset.Client, runtime connector.Runtime) (string, error) {
	pods, err := client.Kubernetes().CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "tier=bfl"})
	if err != nil {
		return "", errors.Wrap(errors.WithStack(err), "get bfl failed")
	}

	if len(pods.Items) == 0 {
		return "", errors.New("bfl not found")
	}

	return pods.Items[0].Spec.NodeName, nil
}

func getDocUrl(ctx context.Context, runtime connector.Runtime) (url string, err error) {
	var nodeip string
	var cmd = fmt.Sprintf(`curl --connect-timeout 30 --retry 5 --retry-delay 1 --retry-max-time 10 -s http://checkip.dyndns.org/ | grep -o "[[:digit:].]\+"`)
	nodeip, _ = runtime.GetRunner().Host.SudoCmdContext(ctx, cmd, false, false)
	url = fmt.Sprintf("http://%s:30883/bfl/apidocs.json", nodeip)
	return
}

func getBflAnnotation(ctx context.Context, ns string, client clientset.Client, runtime connector.Runtime) (map[string]string, error) {
	sts, err := client.Kubernetes().AppsV1().StatefulSets(ns).Get(ctx, "bfl", metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(errors.WithStack(err), "get bfl sts failed")
	}
	if sts == nil {
		return nil, errors.New("bfl sts not found")
	}

	return sts.Annotations, nil
}

func getAppSecrets(patches map[string]map[string]string) map[string]map[string]string {
	var secrests = make(map[string]map[string]string)
	for _, app := range BuiltInApps {
		s, _ := utils.GeneratePassword(16)
		var v = make(map[string]string)
		v["appKey"] = fmt.Sprintf("bytetrade_%s_%d", app, utils.Random())
		v["appSecret"] = s

		p, ok := patches[app]
		if ok && p != nil {
			for pk, pv := range p {
				v[pk] = pv
			}
		}

		secrests[app] = v
	}

	return secrests
}

func getAppPatches() map[string]map[string]string {
	var patches = make(map[string]map[string]string)
	var value = make(map[string]string)
	value["marketProvider"] = os.Getenv("MARKET_PROVIDER")
	patches["market"] = value
	return patches
}
