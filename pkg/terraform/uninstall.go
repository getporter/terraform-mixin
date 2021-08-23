package terraform

import (
	"fmt"

	"get.porter.sh/porter/pkg/exec/builder"
)

// Uninstall runs a terraform destroy
func (m *Mixin) Uninstall() error {
	action, err := m.loadAction()
	if err != nil {
		return err
	}
	step := action.Steps[0]

	err = m.commandPreRun(&step)
	if err != nil {
		return err
	}

	// Update step fields that exec/builder works with
	step.Arguments = []string{"destroy"}
	// Always run in non-interactive mode
	step.Flags = append(step.Flags, builder.NewFlag("auto-approve"))
	step.Flags = append(step.Flags, builder.NewFlag("input=false"))

	for _, k := range sortKeys(step.Vars) {
		step.Flags = append(step.Flags, builder.NewFlag("var", fmt.Sprintf("'%s=%s'", k, step.Vars[k])))
	}

	action.Steps[0] = step
	_, err = builder.ExecuteSingleStepAction(m.Context, action)
	if err != nil {
		return err
	}

	return m.handleOutputs(step.Outputs)
}
