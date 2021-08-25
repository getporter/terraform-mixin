package terraform

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Build(t *testing.T) {
	t.Run("build with the default Terraform version", func(t *testing.T) {
		m := NewTestMixin(t)

		err := m.Build()
		require.NoError(t, err)

		gotOutput := m.TestContext.GetOutput()
		assert.Contains(t, gotOutput, "https://releases.hashicorp.com/terraform/1.0.4/terraform_1.0.4_linux_amd64.zip")
		assert.NotContains(t, "{{.", gotOutput, "Not all of the template values were consumed")
	})

	t.Run("build with custom Terrafrom version", func(t *testing.T) {
		b, err := ioutil.ReadFile("testdata/build-input-with-version.yaml")
		require.NoError(t, err)

		m := NewTestMixin(t)
		m.In = bytes.NewReader(b)
		err = m.Build()
		require.NoError(t, err)

		gotOutput := m.TestContext.GetOutput()
		assert.Contains(t, gotOutput, "https://releases.hashicorp.com/terraform/0.13.0-rc1/terraform_0.13.0-rc1_linux_amd64.zip")
		assert.NotContains(t, "{{.", gotOutput, "Not all of the template values were consumed")
	})
}
