package options

import "github.com/spf13/cobra"

type CliKubeInitializeOptions struct {
	KubeType                    string
	RegistryMirrors             string
	MiniKube                    bool
	MiniKubeProfile             string
	K3sContainerRuntimeEndpoint string
}

func NewCliKubeInitializeOptions() *CliKubeInitializeOptions {
	return &CliKubeInitializeOptions{}
}

func (o *CliKubeInitializeOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().StringVar(&o.MiniKubeProfile, "profile", "", "Set minikube profile name")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().StringVar(&o.K3sContainerRuntimeEndpoint, "k3s-container-runtime-endpoint", "", "Disable k3s embedded containerd and use alternative CRI implementation, e.g., unix:///run/containerd/containerd.sock")
}

type CliTerminusUninstallOptions struct {
	Proxy       string
	MiniKube    bool
	DeleteCRI   bool
	DeleteCache bool
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Set proxy address, e.g., 192.168.50.32 or your-proxy-domain")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().BoolVar(&o.DeleteCRI, "delete-cri", false, "Delete CRI, default: false")
	cmd.Flags().BoolVar(&o.DeleteCache, "delete-cache", false, "Delete Cache, default: false")
}

type CliTerminusInstallOptions struct {
	KubeType string
	Proxy    string
}

func NewCliTerminusInstallOptions() *CliTerminusInstallOptions {
	return &CliTerminusInstallOptions{}
}

func (o *CliTerminusInstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "Set proxy address, e.g., 192.168.50.32 or your-proxy-domain")
}
