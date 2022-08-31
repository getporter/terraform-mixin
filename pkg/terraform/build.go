package terraform

import (
	"text/template"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const dockerfileLines = `
RUN apt-get update && apt-get install -y wget unzip && \
 apt-get clean -y && rm -rf /var/lib/apt/lists/* && \
 wget https://releases.hashicorp.com/terraform/{{.ClientVersion}}/terraform_{{.ClientVersion}}_linux_amd64.zip && \
 unzip terraform_{{.ClientVersion}}_linux_amd64.zip -d /usr/bin && \
 rm terraform_{{.ClientVersion}}_linux_amd64.zip
COPY {{.WorkingDir}}/{{.InitFile}} $BUNDLE_DIR/{{.WorkingDir}}/
RUN cd $BUNDLE_DIR/{{.WorkingDir}} && \
 terraform init -backend=false && \
 rm -fr .terraform/providers && \
 terraform providers mirror /usr/local/share/terraform/plugins
`

// BuildInput represents stdin passed to the mixin for the build command.
type BuildInput struct {
	Config *MixinConfig
}

// MixinConfig represents configuration that can be set on the terraform mixin in porter.yaml
// mixins:
//   - terraform:
//     version: 0.12.17
type MixinConfig struct {
	ClientVersion string `yaml:"clientVersion,omitempty"`
	InitFile      string `yaml:"initFile,omitempty"`
	WorkingDir    string `yaml:"workingDir,omitempty"`
}

func (m *Mixin) Build() error {
	input := BuildInput{
		Config: &m.config, // Apply config directly to the mixin
	}
	err := builder.LoadAction(m.Context, "", func(contents []byte) (interface{}, error) {
		err := yaml.Unmarshal(contents, &input)
		return &input, err
	})
	if err != nil {
		return err
	}

	tmpl, err := template.New("Dockerfile").Parse(dockerfileLines)
	if err != nil {
		return errors.Wrapf(err, "error parsing terraform mixin Dockerfile template")
	}

	return tmpl.Execute(m.Out, input.Config)
}
