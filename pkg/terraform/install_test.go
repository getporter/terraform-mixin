package terraform

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

// sad hack: not sure how to make a common test main for all my subpackages
func TestMain(m *testing.M) {
	test.TestMainWithMockedCommandHandlers(m)
}

func TestMixin_UnmarshalInstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-input.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Install MySQL", step.Description)
	assert.Equal(t, "TRACE", step.LogLevel)
	assert.Equal(t, false, step.Input)
	assert.Equal(t, false, step.DisableVarFile)
}

func TestMixin_Install(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	expectedCommand := strings.Join([]string{
		"terraform init -backend=true -backend-config=key=my.tfstate -reconfigure",
		"terraform apply -auto-approve -input=false -var myvar=foo",
	}, "\n")
	os.Setenv(test.ExpectedCommandEnv, expectedCommand)

	b, err := ioutil.ReadFile("testdata/install-input.yaml")
	require.NoError(t, err)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	// Set up working dir as current dir
	h.WorkingDir = h.Getwd()
	require.NoError(t, err)

	err = h.Install()
	require.NoError(t, err)

	assert.Equal(t, "TRACE", os.Getenv("TF_LOG"))

	wd := h.Getwd()
	require.NoError(t, err)
	assert.Equal(t, wd, h.WorkingDir)
	fc, err := h.FileSystem.ReadFile(path.Join(wd, "terraform.tfvars.json"))
	require.NoError(t, err)
	assert.Equal(t, fc, []byte("{\"myvar\":\"foo\"}"))
}

func TestMixin_UnmarshalInstallSaveVarStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/install-input-disable-save-vars.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Install MySQL", step.Description)
	assert.Equal(t, "TRACE", step.LogLevel)
	assert.Equal(t, false, step.Input)
	assert.Equal(t, true, step.DisableVarFile)
}

func TestMixin_InstallDisableSaveVars(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	expectedCommand := strings.Join([]string{
		"terraform init -backend=true -backend-config=key=my.tfstate -reconfigure",
		"terraform apply -auto-approve -input=false -var myvar=foo",
	}, "\n")
	os.Setenv(test.ExpectedCommandEnv, expectedCommand)

	b, err := ioutil.ReadFile("testdata/install-input-disable-save-vars.yaml")
	require.NoError(t, err)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	// Set up working dir as current dir
	h.WorkingDir = h.Getwd()
	require.NoError(t, err)

	err = h.Install()
	require.NoError(t, err)

	assert.Equal(t, "TRACE", os.Getenv("TF_LOG"))

	wd := h.Getwd()
	assert.Equal(t, wd, h.WorkingDir)
	_, err = h.FileSystem.Stat(path.Join(wd, "terraform.tfvars.json"))
	require.Error(t, err)
}
