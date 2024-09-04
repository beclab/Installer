package options

import "github.com/spf13/cobra"

type CliKubeInitializeOptions struct {
	KubeType        string
	RegistryMirrors string
	MiniKube        bool
	MiniKubeProfile string
}

func NewCliKubeInitializeOptions() *CliKubeInitializeOptions {
	return &CliKubeInitializeOptions{}
}

func (o *CliKubeInitializeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().StringVar(&o.MiniKubeProfile, "profile", "", "Set minikube profile name")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
}

type CliTerminusUninstallOptions struct {
	MiniKube      bool
	StorageType   string // s3 oss
	StorageBucket string

	BaseDir string
	All     bool
	Phase   string

	Quiet bool
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().StringVar(&o.StorageType, "storage-type", "", "Set storage type, e.g., s3 or oss")
	cmd.Flags().StringVar(&o.StorageBucket, "storage-bucket", "", "Set storage bucket")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set uninstall package base dir , default value $HOME/.terminus")
	cmd.Flags().BoolVar(&o.All, "all", false, "Uninstall terminus")
	cmd.Flags().StringVar(&o.Phase, "phase", "prepare", "Uninstall from a specified phase and revert to the previous one. For example, using --phase install will remove the tasks performed in the 'install' phase, effectively returning the system to the 'prepare' state.")

	cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Quiet mode, default: false")
}

type CliTerminusInstallOptions struct {
	Version         string
	KubeType        string
	Proxy           string
	RegistryMirrors string
	MiniKube        bool
	MiniKubeProfile string
	WSL             bool
	BaseDir         string
	Manifest        string

	GpuEnable bool
	GpuShare  bool
}

func NewCliTerminusInstallOptions() *CliTerminusInstallOptions {
	return &CliTerminusInstallOptions{}
}

func (o *CliTerminusInstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Set proxy address, e.g., 192.168.50.32 or your-proxy-domain")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "MacOS platform requires setting minikube parameters, Default: false")
	cmd.Flags().StringVar(&o.MiniKubeProfile, "profile", "", "Set minikube profile name")
	cmd.Flags().BoolVar(&o.WSL, "wsl", false, "Windows platform requires setting WSL parameters, Default: false")
	cmd.Flags().BoolVar(&o.GpuEnable, "gpu-enable", false, "GPU Enable")
	cmd.Flags().BoolVar(&o.GpuShare, "gpu-share", false, "GPU Share")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set pre-install package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value $HOME/.terminus/installation.manifest")
}

type CliPrepareSystemOptions struct {
	Version           string
	KubeType          string
	RegistryMirrors   string
	BaseDir           string
	Manifest          string
	StorageType       string
	StorageDomain     string // s3_bucket --> env.S3_BUCKET
	StorageBucket     string //
	StoragePrefix     string
	StorageAccessKey  string
	StorageSecretKey  string
	StorageToken      string // only for juicefs
	StorageClusterId  string
	StorageSyncSecret string // backup.sync_secret --> env.BACKUP_SECRET, only for cloud
	WSL               bool
}

func NewCliPrepareSystemOptions() *CliPrepareSystemOptions {
	return &CliPrepareSystemOptions{}
}

func (o *CliPrepareSystemOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set pre-install package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value $HOME/.terminus/installation.manifest")
	cmd.Flags().StringVar(&o.StorageType, "storage-type", "minio", "Set storage type, support MinIO, S3, OSS")
	cmd.Flags().StringVar(&o.StorageDomain, "storage-domain", "", "This parameter needs to be set when the storage-type is S3 or OSS, e.g., https://your-name.s3.your-region.amazonaws.com")
	cmd.Flags().StringVar(&o.StorageBucket, "storage-bucket", "", "This parameter needs to be set when the storage-type is S3 or OSS")
	cmd.Flags().StringVar(&o.StoragePrefix, "storage-prefix", "", "This parameter needs to be set when the storage-type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageAccessKey, "storage-access-key", "", "This parameter needs to be set when the storage-type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageSecretKey, "storage-secret-key", "", "This parameter needs to be set when the storage-type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageToken, "storage-token", "", "This parameter needs to be set when the storage-type is S3")
	cmd.Flags().StringVar(&o.StorageClusterId, "cluster-id", "", "")
	cmd.Flags().StringVar(&o.StorageSyncSecret, "sync-secret", "", "")
	cmd.Flags().BoolVar(&o.WSL, "wsl", false, "Windows platform requires setting WSL parameters, Default: false")
}
