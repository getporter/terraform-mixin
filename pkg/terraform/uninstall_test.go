package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type UninstallTest struct {
	expectedCommand string
	uninstallStep   UninstallStep
}

func TestMixin_UnmarshalUninstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	var action UninstallAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Uninstall MySQL", step.Description)
}

func TestMixin_Uninstall(t *testing.T) {
	uninstallTests := []UninstallTest{
		{
			expectedCommand: fmt.Sprintf(
				"terraform destroy -auto-approve -var cool=true -var foo=bar %s", DefaultWorkingDir),
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					AutoApprove: true,
					Vars: map[string]string{
						"cool": "true",
						"foo":  "bar",
					},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, uninstallTest := range uninstallTests {
		t.Run(uninstallTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, uninstallTest.expectedCommand)

			action := UninstallAction{Steps: []UninstallStep{uninstallTest.uninstallStep}}
			b, err := yaml.Marshal(action)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err = h.Uninstall()
			require.NoError(t, err)
		})
	}
}
