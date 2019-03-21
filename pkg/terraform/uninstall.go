package terraform

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// UninstallStep represents the structure of an Uninstall action
type UninstallStep struct {
	UninstallArguments `yaml:"terraform"`
}

// UninstallArguments are the arguments available for the Uninstall action
type UninstallArguments struct {
	Step `yaml:",inline"`
}

// Uninstall runs a terraform destroy
func (m *Mixin) Uninstall() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var step UninstallStep
	err = yaml.Unmarshal(payload, &step)
	if err != nil {
		return err
	}

	cmd := m.NewCommand("terraform", "destroy", "--help")

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

	return nil
}
