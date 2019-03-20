package terraform

import (
	"bufio"
	"io/ioutil"

	"github.com/deislabs/porter/pkg/context"
	"github.com/pkg/errors"
)

// terraform is the logic behind the terraform mixin
type Mixin struct {
	*context.Context
}

// New terraform mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
	}
}

func (m *Mixin) getPayloadData() ([]byte, error) {
	reader := bufio.NewReader(m.In)
	data, err := ioutil.ReadAll(reader)
	return data, errors.Wrap(err, "could not read the payload from STDIN")
}
