package terraform

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"
	"time"

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

func TestMixin_UnmarshalRetry(t *testing.T) {
	b, err := os.ReadFile("testdata/install-input-retry.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)

	assert.Equal(t, &Retry{Attempts: 3, Delay: time.Millisecond}, action.Steps[0].Retry)
}

func TestMixin_UnmarshalRetry_Invalid(t *testing.T) {
	testcases := []struct {
		name    string
		retry   string
		wantErr string
	}{
		{"attempts missing", "delay: 30s", "retry.attempts is required"},
		{"attempts zero", "attempts: 0", "retry.attempts must be at least 1, got 0"},
		{"delay without unit", "attempts: 2\n      delay: 30", `retry.delay must be a duration such as 30s, got "30"`},
		{"delay negative", "attempts: 2\n      delay: -30s", `retry.delay must not be negative, got "-30s"`},
		{"delay empty", "attempts: 2\n      delay: \"\"", `retry.delay must be a duration such as 30s, got ""`},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			input := "install:\n- terraform:\n    description: x\n    retry:\n      " + tc.retry + "\n"
			var action Action
			err := yaml.Unmarshal([]byte(input), &action)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
