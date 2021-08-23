package terraform

import (
	"encoding/json"
	"fmt"
	"path"

	"get.porter.sh/porter/pkg/exec/builder"
)

// defaultTerraformVarFilename is the default name for terrafrom tfvars json file
const defaultTerraformVarsFilename = "terraform.tfvars.json"

// Install runs a terraform apply
func (m *Mixin) Install() error {
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
	step.Arguments = []string{"apply"}
	// Always run in non-interactive mode
	step.Flags = append(step.Flags, builder.NewFlag("auto-approve"))
	step.Flags = append(step.Flags, builder.NewFlag("input=false"))

	// Only create a tf var file for install
	if !step.DisableVarFile && action.Name == "install" {
		vf, err := m.FileSystem.Create(path.Join(m.WorkingDir, defaultTerraformVarsFilename))
		if err != nil {
			return err
		}
		defer vf.Close()

		vbs, err := json.Marshal(step.Vars)
		if err != nil {
			return err
		}

		// If the vars block is empty, set vbs to an empty JSON object
		// to prevent terraform from erroring out
		if len(step.Vars) == 0 {
			vbs = []byte("{}")
		}

		_, err = vf.Write(vbs)
		if err != nil {
			return err
		}

		if m.Debug {
			fmt.Fprintf(m.Err, "DEBUG: TF var file created:\n%s\n", string(vbs))
		}
	}
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
