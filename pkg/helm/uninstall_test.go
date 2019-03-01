package helm

import (
	"bytes"
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

	var step UninstallStep
	err = yaml.Unmarshal(b, &step)
	require.NoError(t, err)

	assert.Equal(t, "Uninstall MySQL", step.Description)
	assert.Equal(t, []string{"porter-ci-mysql"}, step.Releases)
	assert.True(t, step.Purge)
}

func TestMixin_Uninstall(t *testing.T) {
	releases := []string{
		"foo",
		"bar",
	}

	uninstallTests := []UninstallTest{
		{
			expectedCommand: `helm delete foo bar`,
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Releases: releases,
				},
			},
		},
		{
			expectedCommand: `helm delete --purge foo bar`,
			uninstallStep: UninstallStep{
				UninstallArguments: UninstallArguments{
					Purge:    true,
					Releases: releases,
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, uninstallTest := range uninstallTests {
		t.Run(uninstallTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, uninstallTest.expectedCommand)

			b, _ := yaml.Marshal(uninstallTest.uninstallStep)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			err := h.Uninstall()

			require.NoError(t, err)
		})
	}
}
