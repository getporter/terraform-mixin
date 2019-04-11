package main

import (
	"github.com/deislabs/porter-terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func buildUpgradeCommand(m *terraform.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Execute the upgrade functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Upgrade()
		},
	}

	cmd.PersistentFlags().StringVarP(&m.WorkingDir, "work-dir", "w", terraform.DefaultWorkingDir, "The Terraform working directory filepath")

	return cmd
}
