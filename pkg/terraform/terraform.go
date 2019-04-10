package terraform

import (
	"bufio"
	"io/ioutil"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// DefaultWorkingDir is the default working directory for Terraform
const DefaultWorkingDir = "/cnab/app"

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
