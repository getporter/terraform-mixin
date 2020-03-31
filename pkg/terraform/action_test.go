package terraform

import (
	"io/ioutil"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/exec/builder"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
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
	require.Len(t, step.Flags, 4)
	assert.Equal(t, builder.NewFlag("backendConfig", "key=my.tfstate"), step.Flags[0])
	assert.Equal(t, builder.NewFlag("input", "false"), step.Flags[1])
	assert.Equal(t, builder.NewFlag("logLevel", "TRACE"), step.Flags[2])
	assert.Equal(t, builder.NewFlag("vars", "myvar=foo"), step.Flags[3])
}
