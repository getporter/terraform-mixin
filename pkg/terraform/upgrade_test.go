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

type UpgradeTest struct {
	expectedCommand string
	upgradeStep     UpgradeStep
}

func TestMixin_UnmarshalUpgradeStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/upgrade-input.yaml")
	require.NoError(t, err)

	var action UpgradeAction
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Upgrade MySQL", step.Description)
	assert.Equal(t, false, step.Input)
}

func TestMixin_Upgrade(t *testing.T) {
	upgradeTests := []UpgradeTest{
		{
			expectedCommand: strings.Join([]string{
				"terraform init",
				"terraform apply -auto-approve -input=false -var cool=true -var foo=bar",
			}, "\n"),
			upgradeStep: UpgradeStep{
				UpgradeArguments: UpgradeArguments{
					Step: Step{Description: "Upgrade"},
					Vars: map[string]string{
						"cool": "true",
						"foo":  "bar",
					},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, upgradeTest := range upgradeTests {
		t.Run(upgradeTest.expectedCommand, func(t *testing.T) {

			os.Setenv(test.ExpectedCommandEnv, upgradeTest.expectedCommand)
			action := UpgradeAction{Steps: []UpgradeStep{upgradeTest.upgradeStep}}
			b, err := yaml.Marshal(action)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			// Set up working dir as current dir
			h.WorkingDir, err = os.Getwd()
			require.NoError(t, err)

			err = h.Upgrade()
			require.NoError(t, err)

			wd, err := os.Getwd()
			require.NoError(t, err)
			assert.Equal(t, wd, h.WorkingDir)
		})
	}
}
