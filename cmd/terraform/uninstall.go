package main

import (
	"get.porter.sh/mixin/terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func buildUninstallCommand(m *terraform.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Execute the uninstall functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Uninstall()
		},
	}
	return cmd
}
