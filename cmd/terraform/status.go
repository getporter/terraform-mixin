package main

import (
	"github.com/deislabs/porter-terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func buildStatusCommand(m *terraform.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print the status of the terraform components in the bundle",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Status()
		},
	}
	return cmd
}
