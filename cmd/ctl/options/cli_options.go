package options

import (
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
	"github.com/spf13/cobra"
)

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
	Version string
	// KubeType string
	MiniKube bool
	BaseDir  string
	All      bool
	Phase    string
	Quiet    bool
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	// cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "Set minikube flag")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set uninstall package base dir , default value $HOME/.terminus")
	cmd.Flags().BoolVar(&o.All, "all", false, "Uninstall terminus")
	cmd.Flags().StringVar(&o.Phase, "phase", cluster.PhaseInstall.String(), "Uninstall from a specified phase and revert to the previous one. For example, using --phase install will remove the tasks performed in the 'install' phase, effectively returning the system to the 'prepare' state.")
	cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Quiet mode, default: false")
}

type CliTerminusInstallOptions struct {
	Version         string
	KubeType        string
	MiniKube        bool
	MiniKubeProfile string
	BaseDir         string
	Manifest        string
}

func NewCliTerminusInstallOptions() *CliTerminusInstallOptions {
	return &CliTerminusInstallOptions{}
}

func (o *CliTerminusInstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().BoolVar(&o.MiniKube, "minikube", false, "MacOS platform requires setting minikube parameters, Default: false")
	cmd.Flags().StringVar(&o.MiniKubeProfile, "profile", "", "Set minikube profile name")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set pre-install package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value $HOME/.terminus/installation.manifest")
}

type CliPrepareSystemOptions struct {
	Version         string
	KubeType        string
	RegistryMirrors string
	BaseDir         string
	Manifest        string
	Minikube        bool
	MinikubeProfile string
}

func NewCliPrepareSystemOptions() *CliPrepareSystemOptions {
	return &CliPrepareSystemOptions{}
}

func (o *CliPrepareSystemOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "r", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set pre-install package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value $HOME/.terminus/installation.manifest")
	cmd.Flags().BoolVar(&o.Minikube, "minikube", false, "MacOS platform requires setting Minikube parameters, Default: false")
	cmd.Flags().StringVar(&o.MinikubeProfile, "profile", "", "MacOS platform requires setting Minikube Profile parameters")
}
