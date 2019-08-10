package main

import (
	"github.com/deislabs/porter-terraform/pkg/terraform"
	"github.com/spf13/cobra"
)

func buildInvokeCommand(mixin *terraform.Mixin) *cobra.Command {
	opts := terraform.ExecuteCommandOptions{}

	cmd := &cobra.Command{
		Use:   "invoke",
		Short: "Execute the invoke functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return mixin.Execute(opts)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&opts.Action, "action", "", "Custom action name to invoke.")

	return cmd
}
