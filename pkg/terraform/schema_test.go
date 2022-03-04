package terraform

import (
	"io/ioutil"
	"strings"
	"testing"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

func TestMixin_PrintSchema(t *testing.T) {
	m := NewTestMixin(t)

	m.PrintSchema()

	gotSchema := m.TestContext.GetOutput()

	assert.Equal(t, schema, gotSchema)
}

func TestMixin_ValidatePayload(t *testing.T) {
	testcases := []struct {
		name  string
		step  string
		pass  bool
		error string
	}{
		{"install", "testdata/install-input.yaml", true, ""},
		{"install.disable-save-var-file", "testdata/install-input-disable-save-vars.yaml", true, ""},
		{"invoke", "testdata/invoke-input.yaml", true, ""},
		{"upgrade", "testdata/upgrade-input.yaml", true, ""},
		{"uninstall", "testdata/uninstall-input.yaml", true, ""},
		{"install.missing-desc", "testdata/bad-install-input.missing-desc.yaml", false, "install.0.terraform: Invalid type. Expected: object, given: null"},
		{"install.desc-empty", "testdata/bad-install-input.desc-empty.yaml", false, "install.0.terraform.description: String length must be greater than or equal to 1"},
		{"upgrade.disable-var-file", "testdata/bad-upgrade-disable-save-var.yaml", false, "upgrade.0.terraform: Additional property disableVarFile is not allowed"},
		{"uninstall.input-not-valid", "testdata/bad-uninstall-input.input-not-valid.yaml", false, "uninstall.0.terraform: Additional property input is not allowed"},
		{"uninstall.disable-var-file", "testdata/bad-uninstall-disable-save-var.yaml", false, "uninstall.0.terraform: Additional property disableVarFile is not allowed"},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewTestMixin(t)
			b, err := ioutil.ReadFile(tc.step)
			require.NoError(t, err)

			err = m.validatePayload(b)
			if tc.pass {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.error)
			}
		})
	}
}

func (m *Mixin) validatePayload(b []byte) error {
	// Load the step as a go dump
	s := make(map[string]interface{})
	err := yaml.Unmarshal(b, &s)
	if err != nil {
		return errors.Wrap(err, "could not marshal payload as yaml")
	}
	manifestLoader := gojsonschema.NewGoLoader(s)

	// Load the step schema
	schemaLoader := gojsonschema.NewStringLoader(schema)

	validator, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return errors.Wrap(err, "unable to compile the mixin step schema")
	}

	// Validate the manifest against the schema
	result, err := validator.Validate(manifestLoader)
	if err != nil {
		return errors.Wrap(err, "unable to validate the mixin step schema")
	}
	if !result.Valid() {
		errs := make([]string, 0, len(result.Errors()))
		for _, resultErr := range result.Errors() {
			doAppend := true
			for _, err := range errs {
				// no need to append if already exists
				if err == resultErr.String() {
					doAppend = false
				}
			}
			if doAppend {
				errs = append(errs, resultErr.String())
			}
		}
		return errors.New(strings.Join(errs, "\n\t* "))
	}

	return nil
}
