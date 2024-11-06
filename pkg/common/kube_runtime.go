/*
 Copyright 2021 The KubeSphere Authors.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	kubekeyapiv1alpha2 "bytetrade.io/web3os/installer/apis/kubekey/v1alpha2"
	kubekeyclientset "bytetrade.io/web3os/installer/clients/clientset/versioned"
	"bytetrade.io/web3os/installer/pkg/core/connector"
	"bytetrade.io/web3os/installer/pkg/core/logger"
	"bytetrade.io/web3os/installer/pkg/core/storage"
	"bytetrade.io/web3os/installer/pkg/core/util"
)

type KubeRuntime struct {
	connector.BaseRuntime
	ClusterName string
	Cluster     *kubekeyapiv1alpha2.ClusterSpec
	Kubeconfig  string
	ClientSet   *kubekeyclientset.Clientset
	Arg         Argument
}

type Argument struct {
	NodeName            string `json:"node_name"`
	FilePath            string `json:"file_path"`
	KubernetesVersion   string `json:"kubernetes_version"`
	KsEnable            bool   `json:"ks_enable"`
	KsVersion           string `json:"ks_version"`
	TerminusVersion     string `json:"terminus_version"`
	Debug               bool   `json:"debug"`
	IgnoreErr           bool   `json:"ignore_err"`
	SkipPullImages      bool   `json:"skip_pull_images"`
	SKipPushImages      bool   `json:"skip_push_images"`
	SecurityEnhancement bool   `json:"security_enhancement"`
	DeployLocalStorage  *bool  `json:"deploy_local_storage"`
	// DownloadCommand     func(path, url string) string
	SkipConfirmCheck bool   `json:"skip_confirm_check"`
	InCluster        bool   `json:"in_cluster"`
	ContainerManager string `json:"container_manager"`
	FromCluster      bool   `json:"from_cluster"`
	KubeConfig       string `json:"kube_config"`
	Artifact         string `json:"artifact"`
	InstallPackages  bool   `json:"install_packages"`
	ImagesDir        string `json:"images_dir"`
	Namespace        string `json:"namespace"`
	DeleteCRI        bool   `json:"delete_cri"`
	DeleteCache      bool   `json:"delete_cache"`
	Role             string `json:"role"`
	Type             string `json:"type"`
	Kubetype         string `json:"kube_type"`
	SystemInfo       connector.Systems

	// Extra args
	ExtraAddon      string `json:"extra_addon"` // addon yaml config
	RegistryMirrors string `json:"registry_mirrors"`

	// master node ssh config
	MasterHost              string `json:"master_host"`
	MasterNodeName          string `json:"master_node_name"`
	MasterSSHPort           int    `json:"master_ssh_port"`
	MasterSSHUser           string `json:"master_ssh_user"`
	MasterSSHPassword       string `json:"-"`
	MasterSSHPrivateKeyPath string `json:"-"`
	LocalSSHPort            int    `json:"-"`

	SkipMasterPullImages bool `json:"skip_master_pull_images"`

	// db
	Provider storage.Provider `json:"-"`
	// User
	User *User `json:"user"`
	// storage
	Storage                *Storage           `json:"storage"`
	PublicNetworkInfo      *PublicNetworkInfo `json:"public_network_info"`
	GPU                    *GPU               `json:"gpu"`
	Cloudflare             *Cloudflare        `json:"cloudflare"`
	Frp                    *Frp               `json:"frp"`
	TokenMaxAge            int64              `json:"token_max_age"` // nanosecond
	MarketProvider         string             `json:"market_provider"`
	TerminusCertServiceAPI string             `json:"terminus_cert_service_api"`
	TerminusDNSServiceAPI  string             `json:"terminus_dns_service_api"`

	Request any `json:"-"`

	IsCloudInstance    bool     `json:"is_cloud_instance"`
	MinikubeProfile    string   `json:"minikube_profile"`
	WSLDistribution    string   `json:"wsl_distribution"`
	Environment        []string `json:"environment"`
	BaseDir            string   `json:"base_dir"`
	Manifest           string   `json:"manifest"`
	ConsoleLogFileName string   `json:"console_log_file_name"`
	HostIP             string   `json:"host_ip"`
}

type PublicNetworkInfo struct {
	PublicIp string `json:"aws_public_ip"`
	Hostname string `json:"aws_hostname"`
}

type User struct {
	UserName   string `json:"user_name"`
	Password   string `json:"user_password"`
	Email      string `json:"user_email"`
	DomainName string `json:"user_domain_name"`
}

type Storage struct {
	StorageVendor    string `json:"storage_vendor"`
	StorageType      string `json:"storage_type"`
	StorageBucket    string `json:"storage_bucket"`
	StoragePrefix    string `json:"storage_prefix"`
	StorageAccessKey string `json:"storage_access_key"`
	StorageSecretKey string `json:"storage_secret_key"`

	StorageToken        string `json:"storage_token"`       // juicefs  --> from env
	StorageClusterId    string `json:"storage_cluster_id"`  // use only on the Terminus cloud, juicefs  --> from env
	StorageSyncSecret   string `json:"storage_sync_secret"` // use only on the Terminus cloud  --> from env
	BackupClusterBucket string `json:"backup_cluster_bucket"`
}

type GPU struct {
	Enable bool `json:"gpu_enable"`
	Share  bool `json:"gpu_share"`
}

type Cloudflare struct {
	Enable string `json:"cloudflare_enable"`
}

type Frp struct {
	Enable     string `json:"frp_enable"`
	Server     string `json:"frp_server"`
	Port       string `json:"frp_port"`
	AuthMethod string `json:"frp_auth_method"`
	AuthToken  string `json:"frp_auth_token"`
}

func NewArgument() *Argument {
	return &Argument{
		KsEnable:         true,
		KsVersion:        DefaultKubeSphereVersion,
		InstallPackages:  false,
		SKipPushImages:   false,
		ContainerManager: Containerd,
		SystemInfo:       connector.GetSystemInfo(),
		IsCloudInstance:  strings.EqualFold(os.Getenv(ENV_TERMINUS_IS_CLOUD_VERSION), TRUE),
		Storage: &Storage{
			StorageType: Minio,
		},
		GPU: &GPU{
			Enable: strings.EqualFold(os.Getenv(ENV_LOCAL_GPU_ENABLE), "1"),
			Share:  strings.EqualFold(os.Getenv(ENV_LOCAL_GPU_SHARE), "1"),
		},
		Cloudflare:             &Cloudflare{},
		Frp:                    &Frp{},
		User:                   &User{},
		PublicNetworkInfo:      &PublicNetworkInfo{},
		RegistryMirrors:        os.Getenv(ENV_REGISTRY_MIRRORS),
		MarketProvider:         os.Getenv(ENV_MARKET_PROVIDER),
		TerminusCertServiceAPI: os.Getenv(ENV_TERMINUS_CERT_SERVICE_API),
		TerminusDNSServiceAPI:  os.Getenv(ENV_TERMINUS_DNS_SERVICE_API),
		HostIP:                 os.Getenv(ENV_HOST_IP),
		Environment:            os.Environ(),
	}
}

func (a *Argument) GetWslUserPath() string {
	if a.Environment == nil || len(a.Environment) == 0 {
		return ""
	}

	var res string
	var wslSuffix = "/AppData/Local/Microsoft/WindowsApps"
	for _, v := range a.Environment {
		if strings.HasPrefix(v, "PATH=") {
			p := strings.ReplaceAll(v, "PATH=", "")
			s := strings.Split(p, ":")
			for _, s1 := range s {
				if strings.Contains(s1, wslSuffix) {
					res = strings.ReplaceAll(s1, wslSuffix, "")
					break
				}
			}
		}
	}
	return res
}

func (a *Argument) SetTokenMaxAge() {
	s := os.Getenv(ENV_TOKEN_MAX_AGE)
	age, err := strconv.ParseInt(s, 10, 64)
	if err != nil || age == 0 {
		age = DefaultTokenMaxAge
	}
	a.TokenMaxAge = age
}

func (a *Argument) SetGPU(enable bool, share bool) {
	if a.GPU == nil {
		a.GPU = new(GPU)
	}
	a.GPU.Enable = enable
	a.GPU.Share = share
}

func (a *Argument) SetTerminusVersion(version string) {
	if version == "" || len(version) <= 2 {
		return
	}

	if version[0] == 'v' {
		version = version[1:]
	}
	a.TerminusVersion = version
}

func (a *Argument) SetRegistryMirrors(registryMirrors string) {
	a.RegistryMirrors = registryMirrors
}

func (a *Argument) SetDeleteCache(deleteCache bool) {
	a.DeleteCache = deleteCache
}

func (a *Argument) SetDeleteCRI(deleteCRI bool) {
	a.DeleteCRI = deleteCRI
}

func (a *Argument) SetStorage(storage *Storage) {
	a.Storage = storage
}

func (a *Argument) SetMinikubeProfile(profile string) {
	a.MinikubeProfile = profile
	if profile == "" && a.SystemInfo.IsDarwin() {
		fmt.Printf("\nNote: Minikube profile is not set, will try to use the default profile: \"%s\"\n", MinikubeDefaultProfile)
		fmt.Println("if this is not expected, please specify it explicitly by setting the --profile/-p option\n")
		a.MinikubeProfile = MinikubeDefaultProfile
	}
}

func (a *Argument) SetWSLDistribution(distribution string) {
	a.WSLDistribution = distribution
	if distribution == "" && a.SystemInfo.IsWindows() {
		fmt.Printf("\nNote: WSL distribution is not set, will try to use the default distribution: \"%s\"\n", WSLDefaultDistribution)
		fmt.Println("if this is not expected, please specify it explicitly by setting the --distribution/-d option\n")
		a.WSLDistribution = WSLDefaultDistribution
	}
}

func (a *Argument) SetReverseProxy() {
	var enableCloudflare = os.Getenv("CLOUDFLARE_ENABLE")
	var enableFrp = "0"
	var frpServer = ""
	var frpPort = "0"
	var frpAuthMethod = ""
	var frpAuthToken = ""

	if enableCloudflare == "" {
		enableCloudflare = "1"
	}
	if a.IsCloudInstance {
		enableCloudflare = "0"
	} else if os.Getenv("FRP_ENABLE") == "1" {
		enableCloudflare = "0"
		enableFrp = "1"
		frpServer = os.Getenv("FRP_SERVER")
		frpPort = os.Getenv("FRP_PORT")
		frpAuthMethod = os.Getenv("FRP_AUTH_METHOD")
		frpAuthToken = os.Getenv("FRP_AUTH_TOKEN")
	}

	a.Cloudflare.Enable = enableCloudflare
	a.Frp.Enable = enableFrp
	a.Frp.Server = util.RemoveHTTPPrefix(frpServer)
	a.Frp.Port = frpPort
	a.Frp.AuthMethod = frpAuthMethod
	a.Frp.AuthToken = frpAuthToken
}

func (a *Argument) SetKubeVersion(kubeType string) {
	var kubeVersion = DefaultK3sVersion
	if kubeType == K8s {
		kubeVersion = DefaultK8sVersion
	}
	a.KubernetesVersion = kubeVersion
	a.Kubetype = kubeType
}

func (a *Argument) SetKubernetesVersion(kubeType string, kubeVersion string) {
	a.KubernetesVersion = kubeVersion
	a.Kubetype = kubeType
}

func (a *Argument) SetBaseDir(dir string) {
	a.BaseDir = dir
	if !filepath.IsAbs(dir) {
		dir, _ = filepath.Abs(dir)
		if dir != "" {
			a.BaseDir = dir
		}
	}
}

func (a *Argument) SetManifest(manifest string) {
	a.Manifest = manifest
}

func (a *Argument) SetConsoleLogFileName(consoleLogFileName string) {
	a.ConsoleLogFileName = consoleLogFileName
}

func NewKubeRuntime(flag string, arg Argument) (*KubeRuntime, error) {
	loader := NewLoader(flag, arg)
	cluster, err := loader.Load()
	if err != nil {
		return nil, err
	}

	if err = loadExtraAddons(cluster, arg.ExtraAddon); err != nil {
		return nil, err
	}

	base := connector.NewBaseRuntime(cluster.Name, connector.NewDialer(),
		arg.Debug, arg.IgnoreErr, arg.Provider, arg.BaseDir, arg.TerminusVersion, arg.ConsoleLogFileName, arg.SystemInfo)

	clusterSpec := &cluster.Spec
	defaultCluster, roleGroups := clusterSpec.SetDefaultClusterSpec(arg.InCluster, arg.SystemInfo.IsDarwin())
	hostSet := make(map[string]struct{})
	for _, role := range roleGroups {
		for _, host := range role {
			if host.IsRole(Master) || host.IsRole(Worker) {
				host.SetRole(K8s)
			}
			if host.IsRole(Master) && arg.SkipMasterPullImages {
				host.GetCache().Set(SkipMasterNodePullImages, true)
			}
			if _, ok := hostSet[host.GetName()]; !ok {
				hostSet[host.GetName()] = struct{}{}
				base.AppendHost(host)
				base.AppendRoleMap(host)
			}
			host.SetOs(arg.SystemInfo.GetOsType())
			host.SetMinikubeProfile(arg.MinikubeProfile)
		}
	}

	args, _ := json.Marshal(arg)
	logger.Debugf("[runtime] arg: %s", string(args))

	arg.KsEnable = defaultCluster.KubeSphere.Enabled
	arg.KsVersion = defaultCluster.KubeSphere.Version
	r := &KubeRuntime{
		Cluster:     defaultCluster,
		ClusterName: cluster.Name,
		Arg:         arg,
	}
	r.BaseRuntime = base

	return r, nil
}

// Copy is used to create a copy for Runtime.
func (k *KubeRuntime) Copy() connector.Runtime {
	runtime := *k
	return &runtime
}
