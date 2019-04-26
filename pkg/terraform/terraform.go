package terraform

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// DefaultWorkingDir is the default working directory for Terraform
const DefaultWorkingDir = "/cnab/app/terraform"

// terraform is the logic behind the terraform mixin
type Mixin struct {
	*context.Context
	WorkingDir string
}

// New terraform mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context:    context.New(),
		WorkingDir: DefaultWorkingDir,
	}
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	return data, errors.Wrap(err, "could not read the payload from STDIN")
}

func (m *Mixin) getOutput(outputName string) (string, error) {
	cmd := m.NewCommand("terraform", "output", outputName)
	cmd.Stderr = m.Err

	out, err := cmd.Output()
	if err != nil {
		prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
		return "", errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}

	return string(out), nil
}

func (m *Mixin) handleOutputs(outputs []terraformOutput) error {
	var lines []string
	for _, output := range outputs {
		val, err := m.getOutput(output.Name)
		if err != nil {
			return err
		}
		l := fmt.Sprintf("%s=%s", output.Name, val)
		lines = append(lines, l)
	}
	m.Context.WriteOutput(lines)
	return nil
}
