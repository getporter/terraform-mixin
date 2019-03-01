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

	wantOutput := `RUN apt-get update && \
 apt-get install -y curl && \
 curl -o terraform.tgz https://storage.googleapis.com/kubernetes-terraform/terraform-v2.12.3-linux-amd64.tar.gz && \
 tar -xzf terraform.tgz && \
 mv linux-amd64/terraform /usr/local/bin && \
 rm terraform.tgz
RUN terraform init --client-only`

	gotOutput := m.TestContext.GetOutput()
	assert.Equal(t, wantOutput, gotOutput)
}
