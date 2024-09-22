package options

import (
	"github.com/spf13/cobra"
)

type CliDownloadWizardOptions struct {
	Version  string
	KubeType string
	BaseDir  string
}

func NewCliDownloadWizardOptions() *CliDownloadWizardOptions {
	return &CliDownloadWizardOptions{}
}

func (o *CliDownloadWizardOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set download package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
}

type CliDownloadOptions struct {
	Version  string
	KubeType string
	Manifest string
	BaseDir  string
}

func NewCliDownloadOptions() *CliDownloadOptions {
	return &CliDownloadOptions{}
}

func (o *CliDownloadOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.Version, "version", "", "Set install-wizard version, e.g., 1.7.0, 1.7.0-rc.1, 1.8.0-20240813")
	cmd.Flags().StringVar(&o.Manifest, "manifest", "", "Set download package manifest file , default value $HOME/.terminus/installation.manifest")
	cmd.Flags().StringVar(&o.BaseDir, "base-dir", "", "Set download package base dir , default value $HOME/.terminus")
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
}
