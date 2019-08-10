package main

import (
	"fmt"
	"io"
	"os"

	"github.com/deislabs/porter-terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func main() {
	cmd := buildRootCommand(os.Stdin)
	if err := cmd.Execute(); err != nil {
		fmt.Printf("err: %s\n", err)
		os.Exit(1)
	}
}

func buildRootCommand(in io.Reader) *cobra.Command {
	m := terraform.New()
	m.In = in
	cmd := &cobra.Command{
		Use:  "terraform",
		Long: "A terraform mixin for porter 👩🏽‍✈️",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Enable swapping out stdout/stderr for testing
			m.Out = cmd.OutOrStdout()
			m.Err = cmd.OutOrStderr()
		},
		SilenceUsage: true,
	}

	cmd.PersistentFlags().BoolVar(&m.Debug, "debug", false, "Enable debug logging")

	cmd.AddCommand(buildVersionCommand(m))
	cmd.AddCommand(buildSchemaCommand(m))
	cmd.AddCommand(buildBuildCommand(m))
	cmd.AddCommand(buildInstallCommand(m))
	cmd.AddCommand(buildInvokeCommand(m))
	cmd.AddCommand(buildUninstallCommand(m))
	cmd.AddCommand(buildUpgradeCommand(m))
	cmd.AddCommand(buildStatusCommand(m))

	return cmd
}
