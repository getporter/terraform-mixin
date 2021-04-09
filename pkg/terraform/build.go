package terraform

import (
	"fmt"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/ghodss/yaml"
)

const dockerfileLines = `ENV TERRAFORM_VERSION=%s
ENV TERRAFORM_WORKING_DIRECTORY=%s
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
COPY . $BUNDLE_DIR
RUN cd ${TERRAFORM_WORKING_DIRECTORY} && terraform init -backend=false
`

// BuildInput represents stdin passed to the mixin for the build command.
type BuildInput struct {
	Config MixinConfig
}

// MixinConfig represents configuration that can be set on the terraform mixin in porter.yaml
// mixins:
// - terraform:
//	  version: 0.12.17
//    workingDir: terraform
type MixinConfig struct {
	ClientVersion    string `yaml:"clientVersion,omitempty"`
	WorkingDirectory string `yaml:"workingDir,omitempty"`
}

func (m *Mixin) Build() error {
	var input BuildInput
	err := builder.LoadAction(m.Context, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &input)
		return &input, err
	})
	if err != nil {
		return err
	}

	if input.Config.ClientVersion != "" {
		m.TerraformVersion = input.Config.ClientVersion
	}

	if input.Config.WorkingDirectory != "" {
		m.WorkingDir = input.Config.WorkingDirectory
	}

	fmt.Fprintf(m.Out, dockerfileLines, m.TerraformVersion, m.WorkingDir)
	return nil
}
