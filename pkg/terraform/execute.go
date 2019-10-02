package terraform

import (
	"fmt"
	"os"
	"strings"

	"github.com/deislabs/porter/pkg/exec/builder"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

type ExecuteCommandOptions struct {
	Action string
}

type ExecuteAction struct {
	Steps []ExecuteStep // using UnmarshalYAML so that we don't need a custom type per action
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - terraform: ...
// and puts the steps into the Action.Steps field
func (a *ExecuteAction) UnmarshalYAML(unmarshal func(interface{}) error) error {
	actionMap := map[interface{}][]interface{}{}
	err := unmarshal(&actionMap)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal yaml into an action map of terraform steps")
	}

	for _, stepMaps := range actionMap {
		b, err := yaml.Marshal(stepMaps)
		if err != nil {
			return err
		}

		var steps []ExecuteStep
		err = yaml.Unmarshal(b, &steps)
		if err != nil {
			return err
		}

		a.Steps = append(a.Steps, steps...)
	}

	return nil
}

type ExecuteStep struct {
	ExecuteInstruction `yaml:"terraform"`
}

type ExecuteInstruction struct {
	// InstallAguments contains the usual terraform command args for re-use here
	InstallArguments `yaml:",inline"`

	// Command allows an override of the actual terraform command
	Command string `yaml:"command,omitempty"`

	// Flags represents a mapping of a flag to flag value(s) specific to the command
	Flags builder.Flags `yaml:"flags,omitempty"`
}

// Execute will reapply manifests using kubectl
func (m *Mixin) Execute(opts ExecuteCommandOptions) error {

	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action ExecuteAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}

	if len(action.Steps) != 1 {
		return errors.Errorf("expected a single step, but got %d", len(action.Steps))
	}

	step := action.Steps[0]

	if step.LogLevel != "" {
		os.Setenv("TF_LOG", step.LogLevel)
	}

	// First, change to specified working dir
	if err := os.Chdir(m.WorkingDir); err != nil {
		return fmt.Errorf("could not change directory to specified working dir: %s", err)
	}

	// Initialize Terraform
	fmt.Println("Initializing Terraform...")
	err = m.Init(step.BackendConfig)
	if err != nil {
		return fmt.Errorf("could not init terraform, %s", err)
	}

	command := opts.Action
	if step.Command != "" {
		command = step.Command
	}
	cmd := m.NewCommand("terraform", command)

	// All flags in the terraform cli use a single dash
	for i := range step.Flags {
		step.Flags[i].Dash = "-"
	}
	cmd.Args = append(cmd.Args, step.Flags.ToSlice()...)

	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}

	return m.handleOutputs(step.Outputs)
}
