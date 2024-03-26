package terraform

import (
	"io/ioutil"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func TestMixin_UnmarshalStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/step-input.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)

	step := action.Steps[0]
	assert.Equal(t, "Custom Action", step.Description)
	assert.NotEmpty(t, step.Outputs)
	assert.Equal(t, Output{Name: "myoutput"}, step.Outputs[0])

	require.Len(t, step.Arguments, 1)
	assert.Equal(t, "custom", step.Arguments[0])

	sort.Sort(step.Flags)
	require.Len(t, step.Flags, 3)
	assert.Equal(t, builder.NewFlag("backendConfig", "key=my.tfstate"), step.Flags[0])
	assert.Equal(t, builder.NewFlag("logLevel", "TRACE"), step.Flags[1])
	assert.Equal(t, builder.NewFlag("vars", "myvar=foo"), step.Flags[2])
	assert.Equal(t, "testDir", step.WorkingDir)
}

func TestApplyVarsToStepFlags(t *testing.T) {
	t.Run("parse all var data types", func(t *testing.T) {
		s := Step{}
		s.Vars = map[string]interface{}{
			"string": "mystring",
			"bool":   true,
			"int":    22,
			"number": 1.5,
			"list":   []string{"a", "b", "c"},
			"doc": map[string]interface{}{
				"logLevel": "warn",
				"debug":    true,
				"exclude":  []int{1, 2, 3},
				"stuff":    map[string]interface{}{"things": true}},
		}

		applyVarsToStepFlags(&s)

		gotFlags := s.Flags.ToSlice(s.GetDashes())
		wantFlags := []string{
			"-var", `'bool=true'`,
			"-var", `'doc={"debug":true,"exclude":[1,2,3],"logLevel":"warn","stuff":{"things":true}}'`,
			"-var", `'int=22'`,
			"-var", `'list=["a","b","c"]'`,
			"-var", `'number=1.5'`,
			"-var", `'string=mystring'`,
		}
		assert.Equal(t, wantFlags, gotFlags)
	})

	t.Run("empty vars", func(t *testing.T) {
		s := Step{}

		applyVarsToStepFlags(&s)

		gotFlags := s.Flags.ToSlice(s.GetDashes())
		assert.Empty(t, gotFlags)
	})
}

func TestStepGetWorkingDir_ReturnsValidDirectory(t *testing.T) {
	tests := []struct {
		name       string
		workingDir string
		exp        string
	}{
		{
			name:       "Returns . if WorkingDir is empty",
			workingDir: "",
			exp:        ".",
		},
		{
			name:       "Returns value set in WorkingDir",
			workingDir: "testDir",
			exp:        "testDir",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := Step{}
			s.WorkingDir = test.workingDir
			assert.Equal(t, test.exp, s.GetWorkingDir())
		})
	}
}
