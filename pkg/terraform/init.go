package terraform

import (
	"fmt"
	"strings"
)

// Init runs terraform init
func (m *Mixin) Init() error {
	cmd := m.NewCommand("terraform", "init")

	cmd.Stdout = m.Out
	cmd.Stderr = m.Err

	prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
	fmt.Fprintln(m.Out, prettyCmd)

	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("could not execute command, %s: %s", prettyCmd, err)
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
