package main

import (
	"github.com/deislabs/porter-helm/pkg/helm"
	"github.com/spf13/cobra"
)

func buildVersionCommand(m *helm.Mixin) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the mixin version",
		Run: func(cmd *cobra.Command, args []string) {
			m.PrintVersion()
		},
	}
}
