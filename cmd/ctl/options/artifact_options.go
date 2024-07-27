package options

import "github.com/spf13/cobra"

// ~ LoadImageOptions
type LoadImageOptions struct {
	KubeType        string
	RegistryMirrors string
}

func NewLoadImageOptions() *LoadImageOptions {
	return &LoadImageOptions{}
}

func (o *LoadImageOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&o.KubeType, "kube", "k3s", "Set kube type, e.g., k3s or k8s")
	cmd.Flags().StringVarP(&o.RegistryMirrors, "registry-mirrors", "", "", "Docker Container registry mirrors, multiple mirrors are separated by commas")
}
