package terraform

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"

	yaml "gopkg.in/yaml.v2"
)

type ExecuteTest struct {
	expectedCommand string
	executeAction   ExecuteAction
}

func TestMixin_ExecuteStep(t *testing.T) {

	defaultAction := "foo"

	executeTests := []ExecuteTest{
		{
			expectedCommand: fmt.Sprintf("terraform %s", defaultAction),
			executeAction: ExecuteAction{
				Steps: []ExecuteStep{
					ExecuteStep{
						ExecuteInstruction: ExecuteInstruction{
							InstallArguments: InstallArguments{
								Step: Step{
									Description: "My Custom Terraform Action",
								},
							},
						},
					},
				},
			},
		},
		{
			expectedCommand: "terraform version",
			executeAction: ExecuteAction{
				Steps: []ExecuteStep{
					ExecuteStep{
						ExecuteInstruction: ExecuteInstruction{
							Command: "version",
							InstallArguments: InstallArguments{
								Step: Step{
									Description: "My Custom Terraform Action",
								},
							},
						},
					},
				},
			},
		},
	}

	defer os.Unsetenv(test.ExpectedCommandEnv)
	for _, executeTest := range executeTests {
		t.Run(executeTest.expectedCommand, func(t *testing.T) {
			os.Setenv(test.ExpectedCommandEnv, strings.Join([]string{
				"terraform init",
				executeTest.expectedCommand,
			}, "\n"))

			b, err := yaml.Marshal(executeTest.executeAction)
			require.NoError(t, err)

			h := NewTestMixin(t)
			h.In = bytes.NewReader(b)

			// Set up working dir as current dir
			h.WorkingDir, err = os.Getwd()
			require.NoError(t, err)

			err = h.Execute(ExecuteCommandOptions{
				Action: defaultAction,
			})
			require.NoError(t, err)
		})
	}
}
