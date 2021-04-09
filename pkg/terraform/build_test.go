package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const buildOutputTemplate = `ENV TERRAFORM_VERSION=%s
ENV TERRAFORM_WORKING_DIRECTORY=%s
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
COPY . $BUNDLE_DIR
RUN cd ${TERRAFORM_WORKING_DIRECTORY} && terraform init -backend=false
`

const defaultWorkingDirectory = "/cnab/app/terraform"

func TestMixin_Build(t *testing.T) {
	t.Run("build with the default Terraform version", func(t *testing.T) {
		m := NewTestMixin(t)

		err := m.Build()
		require.NoError(t, err)

		wantOutput := fmt.Sprintf(buildOutputTemplate, "0.12.17", defaultWorkingDirectory)

		gotOutput := m.TestContext.GetOutput()
		assert.Equal(t, wantOutput, gotOutput)
	})

	t.Run("build with custom working directory", func(t *testing.T) {
		b, err := ioutil.ReadFile("testdata/build-input-with-working-directory.yaml")
		require.NoError(t, err)

		m := NewTestMixin(t)
		m.In = bytes.NewReader(b)
		err = m.Build()
		require.NoError(t, err)

		expected := fmt.Sprintf(buildOutputTemplate, "0.12.17", "/cnab/app/custom")
		actual := m.TestContext.GetOutput()

		assert.Equal(t, expected, actual)
	})

	t.Run("build with custom Terrafrom version", func(t *testing.T) {
		b, err := ioutil.ReadFile("testdata/build-input-with-version.yaml")
		require.NoError(t, err)

		m := NewTestMixin(t)
		m.In = bytes.NewReader(b)
		err = m.Build()
		require.NoError(t, err)

		wantOutput := fmt.Sprintf(buildOutputTemplate, "0.13.0-rc1", defaultWorkingDirectory)

		gotOutput := m.TestContext.GetOutput()
		assert.Equal(t, wantOutput, gotOutput)
	})
}
