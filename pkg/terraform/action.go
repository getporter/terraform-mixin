package terraform

import (
	"context"
	"encoding/json"
	"fmt"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v2"
)

func (m *Mixin) loadAction(ctx context.Context) (*Action, error) {
	var action Action
	err := builder.LoadAction(ctx, m.RuntimeConfig, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &action)
		return &action, err
	})
	return &action, err
}

var _ builder.ExecutableAction = Action{}
var _ builder.BuildableAction = Action{}

type Action struct {
	Name  string
	Steps []Step // using UnmarshalYAML so that we don't need a custom type per action
}

// MakeSteps builds a slice of Steps for data to be unmarshaled into.
func (a Action) MakeSteps() interface{} {
	return &[]Step{}
}

// UnmarshalYAML takes any yaml in this form
// ACTION:
// - terraform: ...
// and puts the steps into the Action.Steps field
func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	results, err := builder.UnmarshalAction(unmarshal, a)
	if err != nil {
		return err
	}

	for actionName, action := range results {
		a.Name = actionName
		for _, result := range action {
			step := result.(*[]Step)
			a.Steps = append(a.Steps, *step...)
		}
		break // There is only 1 action
	}
	return nil
}

func (a Action) GetSteps() []builder.ExecutableStep {
	// Go doesn't have generics, nothing to see here...
	steps := make([]builder.ExecutableStep, len(a.Steps))
	for i := range a.Steps {
		steps[i] = a.Steps[i]
	}

	return steps
}

type Step struct {
	Instruction `yaml:"terraform"`
}

var _ builder.ExecutableStep = Step{}
var _ builder.HasCustomDashes = Step{}

func (s Step) GetCommand() string {
	return "terraform"
}

func (s Step) GetWorkingDir() string {
	if s.WorkingDir != "" {
		return s.WorkingDir
	}
	return "."
}

func (s Step) GetArguments() []string {
	return s.Arguments
}

func (s Step) GetFlags() builder.Flags {
	return s.Flags
}

func (s Step) GetDashes() builder.Dashes {
	// All flags in the terraform cli use a single dash
	return builder.Dashes{
		Long:  "-",
		Short: "-",
	}
}

func (s Step) GetOutputs() []builder.Output {
	// Go doesn't have generics, nothing to see here...
	outputs := make([]builder.Output, len(s.Outputs))
	for i := range s.Outputs {
		outputs[i] = s.Outputs[i]
	}
	return outputs
}

// applyVarsToStepFlags converts the Terraform vars specified in YAML into a list of -var flags
// with the variable value set in a format that terraform expects.
func applyVarsToStepFlags(step *Step) error {
	if len(step.Vars) == 0 {
		// return early because otherwise parseVars.ForEach below will print `-var =` even when the result is empty
		return nil
	}

	vars, err := json.Marshal(step.Vars)
	if err != nil {
		return fmt.Errorf("error marshaling terraform variables to json")
	}

	parsedVars := gjson.Parse(string(vars))
	parsedVars.ForEach(func(key, value gjson.Result) bool {
		// ensure that the flag value is set using a format that terraform expects
		// primitive data types should print the value directly, e.g. astring, 1, true, 2.4
		// complex data types should be json, e.g. [1,2,3] or {"color":"blue}
		step.Flags = append(step.Flags, builder.NewFlag("var", fmt.Sprintf("'%s=%s'", key.String(), value.String())))
		return true
	})

	return nil
}

type Instruction struct {
	Name            string        `yaml:"name"`
	Description     string        `yaml:"description"`
	Arguments       []string      `yaml:"arguments,omitempty"`
	Flags           builder.Flags `yaml:"flags,omitempty"`
	Outputs         []Output      `yaml:"outputs,omitempty"`
	TerraformFields `yaml:",inline"`
}

// TerraformFields represent fields specific to Terraform
type TerraformFields struct {
	Vars           map[string]interface{} `yaml:"vars,omitempty"`
	DisableVarFile bool                   `yaml:"disableVarFile,omitempty"`
	LogLevel       string                 `yaml:"logLevel,omitempty"`
	BackendConfig  map[string]interface{} `yaml:"backendConfig,omitempty"`
	WorkingDir     string                 `yaml:"workingDir,omitempty"`
}

type Output struct {
	Name string `yaml:"name"`
	// Write the output to the specified file
	DestinationFile string `yaml:"destinationFile,omitempty"`
}

func (o Output) GetName() string {
	return o.Name
}
