package main

import (
	"get.porter.sh/mixin/terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func buildBuildCommand(m *terraform.Mixin) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Generate Dockerfile lines for the bundle invocation image",
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Build(cmd.Context())
		},
	}
	return cmd
}
