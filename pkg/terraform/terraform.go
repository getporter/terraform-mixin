//go:generate packr2

package terraform

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"strings"

	"get.porter.sh/porter/pkg/context" // We are not using go-yaml because of serialization problems with jsonschema, don't use this library elsewhere
	"github.com/gobuffalo/packr/v2"
	"github.com/pkg/errors"
)

// DefaultWorkingDir is the default working directory for Terraform
const DefaultWorkingDir = "/cnab/app/terraform"

// terraform is the logic behind the terraform mixin
type Mixin struct {
	*context.Context
	schema     *packr.Box
	WorkingDir string
}

// New terraform mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context:    context.New(),
		schema:     packr.New("schema", "./schema"),
		WorkingDir: DefaultWorkingDir,
	}
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, errors.Wrap(err, "could not read the payload from STDIN")
	}
	return data, nil
}

func (m *Mixin) getOutput(outputName string) ([]byte, error) {
	cmd := m.NewCommand("terraform", "output", outputName)
	cmd.Stderr = m.Err

	out, err := cmd.Output()
	if err != nil {
		prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}

	return out, nil
}

func (m *Mixin) handleOutputs(outputs []Output) error {
	for _, output := range outputs {
		bytes, err := m.getOutput(output.Name)
		if err != nil {
			return err
		}

		err = m.Context.WriteMixinOutputToFile(output.Name, bytes)
		if err != nil {
			return errors.Wrapf(err, "unable to write output '%s'", output.Name)
		}
	}
	return nil
}
