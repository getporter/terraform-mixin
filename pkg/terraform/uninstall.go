package terraform

import (
	"fmt"
	"os"
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

	// AutoApprove will be deprecated in a later release, it is no longer used, --auto-approve=true is always passed to terraform
	AutoApprove   bool              `yaml:"autoApprove"`
	Vars          map[string]string `yaml:"vars"`
	LogLevel      string            `yaml:"logLevel"`
	BackendConfig map[string]string `yaml:"backendConfig"`
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

	// Run terraform destroy
	cmd := m.NewCommand("terraform", "destroy")

	// Always run in non-interactive mode
	cmd.Args = append(cmd.Args, "-auto-approve")

	for _, k := range sortKeys(step.Vars) {
		cmd.Args = append(cmd.Args, "-var", fmt.Sprintf("%s=%s", k, step.Vars[k]))
	}

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

	m.handleOutputs(step.Outputs)
	return nil
}
