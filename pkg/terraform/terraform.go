package terraform

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"get.porter.sh/porter/pkg/context"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

const (
	// DefaultWorkingDir is the default working directory for Terraform.
	DefaultWorkingDir = "terraform"

	// DefaultClientVersion is the default version of the terraform cli.
	DefaultClientVersion = "1.0.4"

	// DefaultInitFile is the default file used to initialize terraform providers during build.
	DefaultInitFile = ""
)

// Mixin is the logic behind the terraform mixin
type Mixin struct {
	*context.Context
	config MixinConfig
}

// New terraform mixin client, initialized with useful defaults.
func New() *Mixin {
	return &Mixin{
		Context: context.New(),
		config: MixinConfig{
			WorkingDir:    DefaultWorkingDir,
			ClientVersion: DefaultClientVersion,
			InitFile:      DefaultInitFile,
		},
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
	cmd := m.NewCommand("terraform", "output", "-raw", outputName)
	cmd.Stderr = m.Err

	// Terraform appears to auto-append a newline character when printing outputs
	// Trim this character before returning the output
	out, err := cmd.Output()
	out = bytes.TrimRight(out, "\n")

	if err != nil {
		prettyCmd := fmt.Sprintf("%s %s", cmd.Path, strings.Join(cmd.Args, " "))
		return nil, errors.Wrap(err, fmt.Sprintf("couldn't run command %s", prettyCmd))
	}

	return out, nil
}

func (m *Mixin) handleOutputs(outputs []Output) error {
	var bigErr *multierror.Error

	for _, output := range outputs {
		bytes, err := m.getOutput(output.Name)
		if err != nil {
			bigErr = multierror.Append(bigErr, err)
			continue
		}

		err = m.Context.WriteMixinOutputToFile(output.Name, bytes)
		if err != nil {
			bigErr = multierror.Append(bigErr, errors.Wrapf(err, "unable to persist output '%s'", output.Name))
		}

		if output.DestinationFile != "" {
			err = m.Context.FileSystem.MkdirAll(filepath.Dir(output.DestinationFile), 0700)
			if err != nil {
				bigErr = multierror.Append(bigErr, errors.Wrapf(err, "unable to create destination directory for output '%s'", output.Name))
			}

			err = m.Context.FileSystem.WriteFile(output.DestinationFile, bytes, 0700)
			if err != nil {
				bigErr = multierror.Append(bigErr, errors.Wrapf(err, "unable to copy output '%s' to '%s'", output.Name, output.DestinationFile))
			}
		}
	}
	return bigErr.ErrorOrNil()
}

// commandPreRun runs setup tasks applicable for every action
func (m *Mixin) commandPreRun(step *Step) error {
	if step.LogLevel != "" {
		os.Setenv("TF_LOG", step.LogLevel)
	}

	// First, change to specified working dir
	m.Chdir(m.config.WorkingDir)
	if m.Debug {
		fmt.Fprintln(m.Err, "Terraform working directory is", m.Getwd())
	}

	// Initialize Terraform
	fmt.Println("Initializing Terraform...")
	err := m.Init(step.BackendConfig)
	if err != nil {
		return fmt.Errorf("could not init terraform, %s", err)
	}
	return nil
}
