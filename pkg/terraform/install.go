package terraform

import (
	"encoding/json"
	"fmt"
	"path"

	"get.porter.sh/porter/pkg/exec/builder"
)

// DefaultTerraformVarFilename is the default name for terrafrom tfvars json file
const DefaultTerraformVarFilename = "terraform.tfvars.json"

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

	if !step.Input {
		step.Flags = append(step.Flags, builder.NewFlag("input=false"))
	}

	// Only create a tf var file for install
	if !step.DisableVarFile && action.Name == "install" {
		vf, err := m.FileSystem.Create(path.Join(m.WorkingDir, m.TerraformVarsFilename))
		if err != nil {
			return err
		}
		defer vf.Close()
		vbs, err := json.Marshal(step.Vars)
		if err != nil {
			return err
		}
		_, err = vf.Write(vbs)
		if err != nil {
			return err
		}
		if m.Debug {
			fmt.Printf("TF var file created:\n%s\n", string(vbs))
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
