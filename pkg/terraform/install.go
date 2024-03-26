package terraform

import (
	"context"
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/exec/builder"
)

// defaultTerraformVarFilename is the default name for terrafrom tfvars json file
const defaultTerraformVarsFilename = "terraform.tfvars.json"

// Install runs a terraform apply
func (m *Mixin) Install(ctx context.Context) error {
	action, err := m.loadAction(ctx)
	if err != nil {
		return err
	}
	step := action.Steps[0]

	err = m.commandPreRun(ctx, &step)
	if err != nil {
		return err
	}

	// Update step fields that exec/builder works with
	step.Arguments = []string{"apply"}
	// Always run in non-interactive mode
	step.Flags = append(step.Flags, builder.NewFlag("auto-approve"))
	step.Flags = append(step.Flags, builder.NewFlag("input=false"))

	vbs, err := json.Marshal(step.Vars)
	if err != nil {
		return err
	}
	// Only create a tf var file for install
	if !step.DisableVarFile && action.Name == "install" {
		vf, err := m.FileSystem.Create(defaultTerraformVarsFilename)
		if err != nil {
			return err
		}
		defer vf.Close()

		// If the vars block is empty, set vbs to an empty JSON object
		// to prevent terraform from erroring out
		if len(step.Vars) == 0 {
			vbs = []byte("{}")
		}

		_, err = vf.Write(vbs)
		if err != nil {
			return err
		}

		if m.DebugMode {
			fmt.Fprintf(m.Err, "DEBUG: TF var file created:\n%s\n", string(vbs))
		}
	}

	applyVarsToStepFlags(&step)

	action.Steps[0] = step
	_, err = builder.ExecuteSingleStepAction(ctx, m.RuntimeConfig, action)
	if err != nil {
		return err
	}

	return m.handleOutputs(ctx, step.Outputs)
}
