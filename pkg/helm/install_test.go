package helm

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
	assert.NotEmpty(t, step.Outputs)
	assert.Equal(t, HelmOutput{"mysql-root-password", "porter-ci-mysql", "mysql-root-password"}, step.Outputs[0])

	assert.Equal(t, "porter-ci-mysql", step.Name)
	assert.Equal(t, "stable/mysql", step.Chart)
	assert.Equal(t, "0.10.2", step.Version)
	assert.Equal(t, true, step.Replace)
	assert.Equal(t, map[string]string{"mysqlDatabase": "mydb", "mysqlUser": "myuser"}, step.Set)
}

func TestMixin_Install(t *testing.T) {
	namespace := "MYNAMESPACE"
	name := "MYRELEASE"
	chart := "MYCHART"
	version := "1.0.0"
	setArgs := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	values := []string{
		"/tmp/val1.yaml",
		"/tmp/val2.yaml",
	}

	baseInstall := fmt.Sprintf(`helm install --name %s %s --namespace %s --version %s`, name, chart, namespace, version)
	baseValues := `--values /tmp/val1.yaml --values /tmp/val2.yaml`
	baseSetArgs := `--set baz=qux --set foo=bar`

	installTests := []InstallTest{
		{
			expectedCommand: fmt.Sprintf(`%s %s %s`, baseInstall, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseInstall, `--replace`, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Replace:   true,
				},
			},
		},
		{
			expectedCommand: fmt.Sprintf(`%s %s %s %s`, baseInstall, `--wait`, baseValues, baseSetArgs),
			installStep: InstallStep{
				InstallArguments: InstallArguments{
					Namespace: namespace,
					Name:      name,
					Chart:     chart,
					Version:   version,
					Set:       setArgs,
					Values:    values,
					Wait:      true,
				},
			},
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
