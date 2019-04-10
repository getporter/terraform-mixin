package terraform

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

type InstallAction struct {
	Steps []InstallStep `yaml:"install"`
}

// InstallStep represents the structure of an Install action
type InstallStep struct {
	InstallArguments `yaml:"terraform"`
}

// InstallArguments are the arguments available for the Install action
type InstallArguments struct {
	Step `yaml:",inline"`

	AutoApprove bool              `yaml:"autoApprove"`
	Init        bool              `yaml:"init"`
	Vars        map[string]string `yaml:"vars"`
	LogLevel    string            `yaml:"logLevel"`
}

// Install runs a terraform apply
func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var action InstallAction
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
	if step.Init {
		fmt.Println("Initializing Terraform...")
		err = m.Init()
		if err != nil {
			return fmt.Errorf("could not init terraform, %s", err)
		}
	}

	// Next, run Terraform apply
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
