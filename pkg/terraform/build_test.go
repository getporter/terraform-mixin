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
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
COPY . $BUNDLE_DIR
RUN cd /cnab/app/terraform && terraform init -backend=false
`

func TestMixin_Build(t *testing.T) {
	t.Run("build with the default Terraform version", func(t *testing.T) {
		m := NewTestMixin(t)

		err := m.Build()
		require.NoError(t, err)

		wantOutput := fmt.Sprintf(buildOutputTemplate, "0.12.17")

		gotOutput := m.TestContext.GetOutput()
		assert.Equal(t, wantOutput, gotOutput)
	})

	t.Run("build with custom Terrafrom version", func(t *testing.T) {
		b, err := ioutil.ReadFile("testdata/build-input-with-version.yaml")
		require.NoError(t, err)

		m := NewTestMixin(t)
		m.In = bytes.NewReader(b)
		err = m.Build()
		require.NoError(t, err)

		wantOutput := fmt.Sprintf(buildOutputTemplate, "0.13.0-rc1")

		gotOutput := m.TestContext.GetOutput()
		assert.Equal(t, wantOutput, gotOutput)
	})
}
