package terraform

import (
	"fmt"
	"github.com/deislabs/porter/pkg/printer"
	"gopkg.in/yaml.v2"
)

// StatusStep represents the structure of an Status action
type StatusStep struct {
	StatusArguments `yaml:"terraform"`
}

// StatusArguments are the arguments available for the Status action
type StatusArguments struct {
	Step `yaml:",inline"`
}

// Status reports the status for infrastructure provisioned by Terraform
func (m *Mixin) Status(opts printer.PrintOptions) error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var step StatusStep
	err = yaml.Unmarshal(payload, &step)
	if err != nil {
		return err
	}

	format := ""
	switch opts.Format {
	case printer.FormatPlaintext:
		// do nothing, as default output is plaintext
	case printer.FormatYaml:
		format = `-o yaml`
	case printer.FormatJson:
		format = `-o json`
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
	fmt.Sprintf("%s", format)

	return nil
}
