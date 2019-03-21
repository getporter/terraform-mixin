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

	var step InstallStep
	err = yaml.Unmarshal(b, &step)
	require.NoError(t, err)

	assert.Equal(t, "Install MySQL", step.Description)
}

func TestMixin_Install(t *testing.T) {
	installTests := []InstallTest{
		{
			expectedCommand: "terraform apply --help",
			installStep: InstallStep{},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, installTest := range installTests {
		t.Run(installTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, installTest.expectedCommand)

			b, _ := yaml.Marshal(installTest.installStep)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Install()

			require.NoError(t, err)
		})
	}
}
