package options

import (
	"bytetrade.io/web3os/installer/pkg/common"
	"bytetrade.io/web3os/installer/pkg/phase/cluster"
	"github.com/spf13/cobra"
)

type CliTerminusUninstallOptions struct {
	Version string
	BaseDir string
	All     bool
	Phase   string
	Quiet   bool
}

func NewCliTerminusUninstallOptions() *CliTerminusUninstallOptions {
	return &CliTerminusUninstallOptions{}
}

func (o *CliTerminusUninstallOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set uninstall package base dir , default value $HOME/.terminus")
	cmd.Flags().BoolVar(&o.All, "all", false, "Uninstall terminus")
	cmd.Flags().StringVar(&o.Phase, "phase", cluster.PhaseInstall.String(), "Uninstall from a specified phase and revert to the previous one. For example, using --phase install will remove the tasks performed in the 'install' phase, effectively returning the system to the 'prepare' state.")
	cmd.Flags().BoolVar(&o.Quiet, "quiet", false, "Quiet mode, default: false")
}

type CliTerminusInstallOptions struct {
	Version         string
	KubeType        string
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
	cmd.Flags().StringVarP(&o.MiniKubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
	cmd.Flags().StringVarP(&o.BaseDir, "base-dir", "b", "", "Set pre-install package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value {base-dir}/versions/v{version}installation.manifest")
}

type CliPrepareSystemOptions struct {
	Version         string
	KubeType        string
	RegistryMirrors string
	BaseDir         string
	Manifest        string
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
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set pre-install package manifest file , default value {base-dir}/versions/v{version}installation.manifest")
	cmd.Flags().StringVarP(&o.MinikubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
}

type ChangeIPOptions struct {
	Version         string
	BaseDir         string
	WSLDistribution string
	MinikubeProfile string
}

func NewChangeIPOptions() *ChangeIPOptions {
	return &ChangeIPOptions{}
}

func (o *ChangeIPOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.Version, "version", "v", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set uninstall package base dir , defaults to $HOME/.terminus")
	cmd.Flags().StringVarP(&o.WSLDistribution, "distribution", "d", "", "Set WSL distribution name, only in Windows platform, defaults to "+common.WSLDefaultDistribution)
	cmd.Flags().StringVarP(&o.MinikubeProfile, "profile", "p", "", "Set Minikube profile name, only in MacOS platform, defaults to "+common.MinikubeDefaultProfile)
}
