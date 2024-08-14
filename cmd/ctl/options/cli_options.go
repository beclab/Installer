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
	Proxy         string
	MiniKube      bool
	DeleteCRI     bool
	DeleteCache   bool
	StorageType   string // s3 oss
	StorageBucket string
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Set proxy address, e.g., 192.168.50.32 or your-proxy-domain")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().BoolVar(&o.DeleteCRI, "delete-cri", false, "Delete CRI, default: false")
	cmd.Flags().BoolVar(&o.DeleteCache, "delete-cache", false, "Delete Cache, default: false")

	cmd.Flags().StringVar(&o.StorageType, "storage-type", "", "Set storage type, e.g., s3 or oss")
	cmd.Flags().StringVar(&o.StorageBucket, "storage-bucket", "", "Set storage bucket")
}

type CliTerminusInstallOptions struct {
	KubeType         string
	Proxy            string
	RegistryMirrors  string
	MiniKube         bool
	MiniKubeProfile  string
	StorageType      string
	StorageBucket    string
	StorageAccessKey string
	StorageSecretKey string
	StorageToken     string
}

func NewCliTerminusInstallOptions() *CliTerminusInstallOptions {
	return &CliTerminusInstallOptions{}
}

func (o *CliTerminusInstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Set proxy address, e.g., 192.168.50.32 or your-proxy-domain")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().StringVar(&o.MiniKubeProfile, "profile", "", "Set minikube profile name")
	cmd.Flags().StringVar(&o.StorageType, "storage-type", "minio", "Set storage type, support MinIO, S3, OSS")
	cmd.Flags().StringVar(&o.StorageBucket, "storage-bucket", "", "This parameter needs to be set when the storage type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageAccessKey, "storage-access-key", "", "This parameter needs to be set when the storage type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageSecretKey, "storage-secret-key", "", "This parameter needs to be set when the storage type is S3 or OSS")
	cmd.Flags().StringVar(&o.StorageToken, "storage-token", "", "This parameter needs to be set when the storage type is S3 or OSS")
}
