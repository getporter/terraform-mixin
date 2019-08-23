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
	"gopkg.in/yaml.v2"
)

type InstallTest struct {
	expectedCommand string
	installStep     InstallStep
}

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_UnmarshalInstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-input.yaml")
	require.NoError(t, err)

	var action InstallAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Install MySQL", step.Description)
	assert.Equal(t, "TRACE", step.LogLevel)
	assert.Equal(t, false, step.Input)
}

func TestMixin_Install(t *testing.T) {
	installTests := []InstallTest{
		{
			expectedCommand: strings.Join([]string{
				"terraform init",
				"terraform apply -auto-approve -input=false -var cool=true -var foo=bar",
			}, "\n"),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Step: Step{Description: "Install"},
					AutoApprove: true,
					LogLevel:    "TRACE",
					Vars: map[string]string{
						"cool": "true",
						"foo":  "bar",
					},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, installTest := range installTests {
		t.Run(installTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

			action := InstallAction{Steps: []InstallStep{installTest.installStep}}
			b, err := yaml.Marshal(action)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			// Set up working dir as current dir
			h.WorkingDir, err = os.Getwd()
			require.NoError(t, err)

			err = h.Install()
			require.NoError(t, err)

			assert.Equal(t, "TRACE", os.Getenv("TF_LOG"))

			wd, err := os.Getwd()
			require.NoError(t, err)
			assert.Equal(t, wd, h.WorkingDir)
		})
	}
}
