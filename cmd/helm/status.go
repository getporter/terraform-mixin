package main

import (
	"github.com/deislabs/porter-helm/pkg/helm"
	"github.com/deislabs/porter/pkg/printer"
	"github.com/spf13/cobra"
)

func buildStatusCommand(m *helm.Mixin) *cobra.Command {
	opts := struct {
		rawFormat string
		format    printer.Format
	}{}
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print the status of the helm components in the bundle",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			opts.format, err = printer.ParseFormat(opts.rawFormat)
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return m.Status(printer.PrintOptions{Format: opts.format})
		},
	}

	cmd.Flags().StringVarP(&opts.rawFormat, "output", "o", "plaintext", "Output format. Allowed values are: plaintext, yaml, json")
	return cmd
}
