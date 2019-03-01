package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/printer"
	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type statusTest struct {
	format                printer.Format
	expectedCommandSuffix string
}

func TestMixin_UnmarshalStatusStep(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/status-input.yaml")
	require.NoError(t, err)

	var step StatusStep
	err = yaml.Unmarshal(b, &step)
	require.NoError(t, err)

	assert.Equal(t, "Status MySQL", step.Description)
	assert.Equal(t, []string{"porter-ci-mysql"}, step.Releases)
}

func TestMixin_Status(t *testing.T) {
	testCases := map[string]statusTest{
		"default": {
			format:                printer.FormatPlaintext,
			expectedCommandSuffix: "",
		},
		"json": {
			format:                printer.FormatJson,
			expectedCommandSuffix: "-o json",
		},
		"yaml": {
			format:                printer.FormatYaml,
			expectedCommandSuffix: "-o yaml",
		},
	}

	releases := []string{
		"foo",
		"bar",
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for testName, testCase := range testCases {
		for _, release := range releases {
			t.Run(testName, func(t *testing.T) {
				os.Setenv(test.ExpectedCommandEnv,
					strings.TrimSpace(fmt.Sprintf(`terraform status %s %s`, release, testCase.expectedCommandSuffix)))

				statusStep := StatusStep{
					StatusArguments: StatusArguments{
						Releases: []string{release},
					},
				}

				b, _ := yaml.Marshal(statusStep)

				h := NewTestMixin(t)
				h.In = bytes.NewReader(b)

				opts := printer.PrintOptions{testCase.format}
				err := h.Status(opts)

				require.NoError(t, err)
			})
		}
	}
}
