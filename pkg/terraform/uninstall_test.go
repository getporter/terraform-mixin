package terraform

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestMixin_UnmarshalUninstallStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	var action Action
	err = yaml.Unmarshal(b, &action)
	require.NoError(t, err)
	require.Len(t, action.Steps, 1)
	step := action.Steps[0]

	assert.Equal(t, "Uninstall MySQL", step.Description)
}

func TestMixin_Uninstall(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	expectedCommand := strings.Join([]string{
		"terraform init -backend=true -backend-config=key=my.tfstate -reconfigure",
		"terraform destroy -auto-approve -input=false -var myvar=foo",
	}, "\n")
	os.Setenv(test.ExpectedCommandEnv, expectedCommand)

	b, err := ioutil.ReadFile("testdata/uninstall-input.yaml")
	require.NoError(t, err)

	h := NewTestMixin(t)
	h.In = bytes.NewReader(b)

	// Set up working dir as current dir
	h.config.WorkingDir = h.Getwd()
	require.NoError(t, err)

	err = h.Uninstall(context.Background())
	require.NoError(t, err)

	wd := h.Getwd()
	require.NoError(t, err)
	assert.Equal(t, wd, h.config.WorkingDir)
}
