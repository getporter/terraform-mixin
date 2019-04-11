package main

import (
	"github.com/deislabs/porter-terraform/pkg/terraform"
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
			return m.Install()
		},
	}

	cmd.PersistentFlags().StringVarP(&m.WorkingDir, "work-dir", "w", terraform.DefaultWorkingDir, "The Terraform working directory filepath")

	return cmd
}
