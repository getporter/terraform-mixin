package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Build(t *testing.T) {
	m := NewTestMixin(t)

	err := m.Build()
	require.NoError(t, err)

	wantOutput := `ENV TERRAFORM_VERSION=0.11.11
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin`

	gotOutput := m.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}
