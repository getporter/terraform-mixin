package terraform

import (
	"fmt"
	"os"

	"get.porter.sh/porter/pkg/exec/builder"
)

type InvokeOptions struct {
	Action string
}

// Invoke runs a custom terraform action
func (m *Mixin) Invoke(opts InvokeOptions) error {
	action, err := m.loadAction()
	if err != nil {
		return err
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

	// Update step fields that exec/builder works with
	commands := []string{opts.Action}
	if len(step.Arguments) > 0 {
		commands = step.GetArguments()
	}
	step.Arguments = commands

	for _, k := range sortKeys(step.Vars) {
		step.Flags = append(step.Flags, builder.NewFlag("var", fmt.Sprintf("%s=%s", k, step.Vars[k])))
	}

	action.Steps[0] = step
	_, err = builder.ExecuteSingleStepAction(m.Context, action)
	if err != nil {
		return err
	}

	return m.handleOutputs(step.Outputs)
}
