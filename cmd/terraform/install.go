package main

import (
	"get.porter.sh/mixin/terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

var (
	commandFile string
)

func buildInstallCommand(m *terraform.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Execute the install functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Install(cmd.Context())
		},
	}
	return cmd
}
