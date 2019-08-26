package terraform

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
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
			expectedCommand: strings.Join([]string{
				"terraform init",
				"terraform destroy -auto-approve -var cool=true -var foo=bar",
			}, "\n"),
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Step: Step{Description: "Uninstall"},
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

			// Set up working dir as current dir
			h.WorkingDir, err = os.Getwd()
			require.NoError(t, err)

			err = h.Uninstall()
			require.NoError(t, err)

			wd, err := os.Getwd()
			require.NoError(t, err)
			assert.Equal(t, wd, h.WorkingDir)
		})
	}
}
