package terraform

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"get.porter.sh/mixin/terraform/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMixin_Build(t *testing.T) {
	testcases := []struct {
		name                    string
		inputFile               string
		expectedVersion         string
		expectedUserAgent       string
		expectedDockerfileLines string
	}{
		{
			name:                    "build with custom config",
			inputFile:               "testdata/build-input-with-config.yaml",
			expectedVersion:         "https://releases.hashicorp.com/terraform/0.13.0-rc1/terraform_0.13.0-rc1_linux_amd64.zip",
			expectedUserAgent:       "ENV PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT=\"true\"\nENV AZURE_HTTP_USER_AGENT=\"\"",
			expectedDockerfileLines: "COPY terraform/ $BUNDLE_DIR/terraform/\nRUN cd $BUNDLE_DIR/terraform && \\\n terraform init -backend=false && \\\n rm -fr .terraform/providers && \\\n terraform providers mirror /usr/local/share/terraform/plugins\n\n",
		},
		{
			name:                    "build with the default Terraform config",
			expectedVersion:         "https://releases.hashicorp.com/terraform/1.2.9/terraform_1.2.9_linux_amd64.zip",
			expectedUserAgent:       "ENV PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT=\"false\"\nENV AZURE_HTTP_USER_AGENT=\"getporter/porter getporter/terraform/v1.2.3",
			expectedDockerfileLines: "COPY terraform/ $BUNDLE_DIR/terraform/\nRUN cd $BUNDLE_DIR/terraform && \\\n terraform init -backend=false && \\\n rm -fr .terraform/providers && \\\n terraform providers mirror /usr/local/share/terraform/plugins\n\n",
		},
		{
			name:                    "build with mulitple terraform working dirs",
			inputFile:               "testdata/build-input-with-multiple-working-dirs.yaml",
			expectedVersion:         "https://releases.hashicorp.com/terraform/1.5.0/terraform_1.5.0_linux_amd64.zip",
			expectedUserAgent:       "ENV PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT=\"true\"\nENV AZURE_HTTP_USER_AGENT=\"\"",
			expectedDockerfileLines: "COPY testDir1/ $BUNDLE_DIR/testDir1/\nRUN cd $BUNDLE_DIR/testDir1 && \\\n terraform init -backend=false && \\\n rm -fr .terraform/providers && \\\n terraform providers mirror /usr/local/share/terraform/plugins\n\nCOPY testDir2/ $BUNDLE_DIR/testDir2/\nRUN cd $BUNDLE_DIR/testDir2 && \\\n terraform init -backend=false && \\\n rm -fr .terraform/providers && \\\n terraform providers mirror /usr/local/share/terraform/plugins\n\n\n",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Set a fake version of the mixin and porter for our user agent
			pkg.Version = "v1.2.3"

			var data []byte
			var err error
			if tc.inputFile != "" {
				data, err = ioutil.ReadFile(tc.inputFile)
				require.NoError(t, err)
			}

			m := NewTestMixin(t)
			m.In = bytes.NewReader(data)

			err = m.Build(context.Background())
			require.NoError(t, err, "build failed")

			gotOutput := m.TestContext.GetOutput()
			assert.Contains(t, gotOutput, tc.expectedVersion)
			assert.Contains(t, gotOutput, tc.expectedUserAgent)
			assert.NotContains(t, "{{.", gotOutput, "Not all of the template values were consumed")
			assert.Contains(t, gotOutput, tc.expectedDockerfileLines)
		})
	}
}
