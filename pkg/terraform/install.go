package terraform

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

// InstallStep represents the structure of an Install action
type InstallStep struct {
	InstallArguments `yaml:"terraform"`


}

// InstallArguments are the arguments available for the Install action
type InstallArguments struct {
	Step `yaml:",inline"`
}

// Install runs a terraform apply
func (m *Mixin) Install() error {
	payload, err := m.getPayloadData()
	if err != nil {
		return err
	}

	var step InstallStep
	err = yaml.Unmarshal(payload, &step)
	if err != nil {
		return err
	}

	cmd := m.NewCommand("terraform", "apply", "--help")

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
