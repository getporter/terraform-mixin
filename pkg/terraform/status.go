package terraform

import (
	"fmt"
	"strings"

	"github.com/deislabs/porter/pkg/printer"
	yaml "gopkg.in/yaml.v2"
)

// StatusStep represents the structure of an Status action
type StatusStep struct {
	StatusArguments `yaml:"terraform"`
}

// StatusArguments are the arguments available for the Status action
type StatusArguments struct {
	Step `yaml:",inline"`

	Releases []string `yaml:"releases"`
}

// Status reports the status for a provided set of terraform releases
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

	for _, release := range step.Releases {
		cmd := m.NewCommand("terraform", "status", strings.TrimSpace(fmt.Sprintf(`%s %s`, release, format)))

		cmd.Stdout = m.Out
		cmd.Stderr = m.Err

		prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
		fmt.Fprintln(m.Out, prettyCmd)

		err = cmd.Start()
		if err != nil {
			return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
		}
		err = cmd.Wait()
		if err != nil {
			return err
		}
	}

	return nil
}
