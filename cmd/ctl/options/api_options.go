package options

import "github.com/spf13/cobra"

type ApiOptions struct {
	Enabled bool
	Port    string
	Proxy   string
}

func NewApiOptions() *ApiOptions {
	return &ApiOptions{}
}

func (o *ApiOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.Enabled, "enabled", false, "running api server")
	cmd.Flags().StringVar(&o.Port, "port", ":30080", "listen port")
	cmd.Flags().StringVar(&o.Proxy, "proxy", "", "proxy")
}
