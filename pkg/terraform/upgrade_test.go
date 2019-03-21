package terraform

import (
	"bytes"
	"io/ioutil"
	"os"
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

	var step UpgradeStep
	err = yaml.Unmarshal(b, &step)
	require.NoError(t, err)

	assert.Equal(t, "Upgrade MySQL", step.Description)
}

func TestMixin_Upgrade(t *testing.T) {
	upgradeTests := []UpgradeTest{
		{
			expectedCommand: "terraform apply --help",
			upgradeStep: UpgradeStep{},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, upgradeTest := range upgradeTests {
		t.Run(upgradeTest.expectedCommand, func(t *testing.T) {

			os.Setenv(test.ExpectedCommandEnv, upgradeTest.expectedCommand)
			b, err := yaml.Marshal(upgradeTest.upgradeStep)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err = h.Upgrade()

			require.NoError(t, err)
		})
	}
}
