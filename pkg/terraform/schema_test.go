package terraform

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/PaesslerAG/jsonpath"
	"github.com/ghodss/yaml"
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

func TestMixin_CheckSchema(t *testing.T) {
	// Long term it would be great to have a helper function in Porter that a mixin can use to check that it meets certain interfaces
	// check that certain characteristics of the schema that Porter expects are present
	// Once we have a mixin library, that would be a good place to package up this type of helper function
	var schemaMap map[string]interface{}
	err := json.Unmarshal([]byte(schema), &schemaMap)
	require.NoError(t, err, "could not unmarshal the schema into a map")

	t.Run("mixin configuration", func(t *testing.T) {
		// Check that mixin config is defined, and has all the supported fields
		configSchema, err := jsonpath.Get("$.definitions.config", schemaMap)
		require.NoError(t, err, "could not find the mixin config schema declaration")
		_, err = jsonpath.Get("$.properties.terraform.properties.clientVersion", configSchema)
		require.NoError(t, err, "clientVersion was not included in the mixin config schema")
		_, err = jsonpath.Get("$.properties.terraform.properties.initFile", configSchema)
		require.NoError(t, err, "initFile was not included in the mixin config schema")
		_, err = jsonpath.Get("$.properties.terraform.properties.workingDir", configSchema)
		require.NoError(t, err, "workingDir was not included in the mixin config schema")
	})

	// Check that schema are defined for each action
	actions := []string{"install", "upgrade", "invoke", "uninstall"}
	for _, action := range actions {
		t.Run("supports "+action, func(t *testing.T) {
			actionPath := fmt.Sprintf("$.definitions.%sStep", action)
			_, err := jsonpath.Get(actionPath, schemaMap)
			require.NoErrorf(t, err, "could not find the %sStep declaration", action)
		})
	}

	// Check that the invoke action is registered
	additionalSchema, err := jsonpath.Get("$.additionalProperties.items", schemaMap)
	require.NoError(t, err, "the invoke action was not registered in the schema")
	require.Contains(t, additionalSchema, "$ref")
	invokeRef := additionalSchema.(map[string]interface{})["$ref"]
	require.Equal(t, "#/definitions/invokeStep", invokeRef, "the invoke action was not registered correctly")
}
