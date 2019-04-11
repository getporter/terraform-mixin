package main

import (
	"github.com/deislabs/porter-terraform/pkg/terraform"
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

	cmd.PersistentFlags().StringVarP(&m.WorkingDir, "work-dir", "w", terraform.DefaultWorkingDir, "The Terraform working directory filepath")

	return cmd
}
