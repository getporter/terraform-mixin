package terraform

import (
	"fmt"
	"strings"
)

// Init runs terraform init with the provided backendConfig, if supplied
func (m *Mixin) Init(backendConfig map[string]interface{}) error {
	cmd := m.NewCommand("terraform", "init")

	if len(backendConfig) > 0 {
		cmd.Args = append(cmd.Args, "-backend=true")

		for _, k := range sortKeys(backendConfig) {
			cmd.Args = append(cmd.Args, fmt.Sprintf("-backend-config=%s=%s", k, backendConfig[k]))
		}

		cmd.Args = append(cmd.Args, "-reconfigure")
	}

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
