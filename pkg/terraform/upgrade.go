package terraform

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type UpgradeAction struct {
	Steps []UpgradeStep `yaml:"upgrade"`
}

// UpgradeStep represents the structure of an Upgrade step
type UpgradeStep struct {
	UpgradeArguments `yaml:"terraform"`
}

// UpgradeArguments represent the arguments available to the Upgrade step
type UpgradeArguments struct {
	Step `yaml:",inline"`

	AutoApprove bool              `yaml:"autoApprove"`
	Vars        map[string]string `yaml:"vars"`
	LogLevel    string            `yaml:"logLevel"`
}

// Upgrade runs a terraform apply
func (m *Mixin) Upgrade() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action UpgradeAction
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
	err = m.Init()
	if err != nil {
		return fmt.Errorf("could not init terraform, %s", err)
	}

	// Run terraform apply
	cmd := m.NewCommand("terraform", "apply")

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
