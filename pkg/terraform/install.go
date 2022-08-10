package terraform

import (
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/tidwall/gjson"
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
		vf, err := m.FileSystem.Create(defaultTerraformVarsFilename)
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
	v, err := json.Marshal(step.Vars)
	if err != nil {
		return err
	}
	//`{"var1": "foo", "var2": ["bar", "baz"], "var3": {"biz": "box"}`
	result := gjson.Parse(string(v))
	result.ForEach(func(key, value gjson.Result) bool {
		step.Flags = append(step.Flags, builder.NewFlag("var", fmt.Sprintf("'%s=%s'", key.String(), value.String())))
		return true
	})
	// gjson.ForEachLine(string(v), func(line gjson.Result) bool {
	// 	return false
	// })
	// for _, k := range sortKeys(step.Vars) {
	// 	//fmt.Printf("\n\n\n\nSETTING VARS FILES\n\n\n\n\n")
	// 	v, err := json.Marshal(step.Vars)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	val := gjson.GetBytes(v, k)
	// 	str := val.String()
	// 	//v, err := json.Marshal(step.Vars[k])
	// 	//v, err := convertValue(step.Vars[k])
	// 	if err != nil {
	// 		fmt.Printf("\n\n\n\nERR: %s\n\n\n\n", err.Error())
	// 	}
	// 	//fmt.Printf("\n\n\n\nV: %s\n\n\n\n", string(v))
	// 	step.Flags = append(step.Flags, builder.NewFlag("var", fmt.Sprintf("'%s=%s'", k, str)))
	// }
	//step.Flags = append(step.Flags, builder.NewFlag("var-file", defaultTerraformVarsFilename))
	//fmt.Printf("\n\nFLAGS: %#v\n\n", step.Flags)
	action.Steps[0] = step
	_, err = builder.ExecuteSingleStepAction(m.Context, action)
	if err != nil {
		return err
	}

	return m.handleOutputs(step.Outputs)
}
