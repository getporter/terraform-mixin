package terraform

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type StatusAction struct {
	Steps []StatusStep `yaml:"status"`
}

// StatusStep represents the structure of an Status action
type StatusStep struct {
	StatusArguments `yaml:"terraform"`
}

// StatusArguments are the arguments available for the Status action
type StatusArguments struct {
	Step `yaml:",inline"`

	LogLevel      string            `yaml:"logLevel"`
	BackendConfig map[string]string `yaml:"backendConfig"`
}

// Status reports the status for infrastructure provisioned by Terraform
func (m *Mixin) Status() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action StatusAction
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

	// Run terraform show
	cmd := m.NewCommand("terraform", "show")

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
