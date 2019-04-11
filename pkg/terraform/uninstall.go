package terraform

import (
	"fmt"
	"os"
	"sort"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type UninstallAction struct {
	Steps []UninstallStep `yaml:"uninstall"`
}

// UninstallStep represents the structure of an Uninstall action
type UninstallStep struct {
	UninstallArguments `yaml:"terraform"`
}

// UninstallArguments are the arguments available for the Uninstall action
type UninstallArguments struct {
	Step `yaml:",inline"`

	AutoApprove bool              `yaml:"autoApprove"`
	Vars        map[string]string `yaml:"vars"`
	LogLevel    string            `yaml:"logLevel"`
}

// Uninstall runs a terraform destroy
func (m *Mixin) Uninstall() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action UninstallAction
	err = yaml.Unmarshal(payload, &action)
	if err != nil {
		return err
	}
	if len(action.Steps) != 1 {
		return fmt.Errorf("expected a single step, but got %d", len(action.Steps))
	}
	step := action.Steps[0]

	if step.LogLevel != "" {
		os.Setenv("TF_LOG", step.LogLevel)
	}

	// First, initialize Terraform
	fmt.Println("Initializing Terraform...")
	err = m.Init()
	if err != nil {
		return fmt.Errorf("could not init terraform, %s", err)
	}

	// Next, run terraform destroy
	cmd := m.NewCommand("terraform", "destroy")

	if step.AutoApprove {
		cmd.Args = append(cmd.Args, "-auto-approve")
	}

	// sort the vars consistently
	varKeys := make([]string, 0, len(step.Vars))
	for k := range step.Vars {
		varKeys = append(varKeys, k)
	}
	sort.Strings(varKeys)

	for _, k := range varKeys {
		cmd.Args = append(cmd.Args, "-var", fmt.Sprintf("%s=%s", k, step.Vars[k]))
	}

	// Configuration path must represent the last argument
	cmd.Args = append(cmd.Args, m.WorkingDir)

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

	return nil
}
