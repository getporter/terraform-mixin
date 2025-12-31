package terraform

import (
	"context"
	"text/template"

	"get.porter.sh/porter/pkg/exec/builder"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

const dockerfileLines = `
ENV PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT="{{ .UserAgentOptOut}}"
ENV AZURE_HTTP_USER_AGENT="{{ .AzureUserAgent }}"
RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt \
 apt-get update && apt-get install -y wget unzip && \
 wget {{or .InstallHost "https://releases.hashicorp.com"}}/terraform/{{.ClientVersion}}/terraform_{{.ClientVersion}}_linux_amd64.zip --progress=dot:giga && \
 unzip terraform_{{.ClientVersion}}_linux_amd64.zip -d /usr/bin && \
 rm terraform_{{.ClientVersion}}_linux_amd64.zip
COPY {{.WorkingDir}}/{{.InitFile}} $BUNDLE_DIR/{{.WorkingDir}}/
{{if .ProviderHost }}
ENV TF_CLI_CONFIG_FILE=$BUNDLE_DIR/{{.WorkingDir}}/provider_mirror.tfrc
RUN tee <<EOF > $TF_CLI_CONFIG_FILE
  provider_installation {
      direct {
          exclude = ["registry.terraform.io/*/*"]
      }
      network_mirror {
          url = "{{ .ProviderHost }}"
      }
  }
EOF
{{ end }}
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
	// ClientVersion is the version of the terraform CLI to install
	ClientVersion string `yaml:"clientVersion,omitempty"`

	// UserAgentOptOut allows a bundle author to opt out from adding porter and the mixin's version to the terraform user agent string.
	UserAgentOptOut bool `yaml:"userAgentOptOut,omitempty"`

	InitFile   string `yaml:"initFile,omitempty"`
	WorkingDir string `yaml:"workingDir,omitempty"`

	// Host from which to install `terraform`.
	InstallHost string `yaml:"installHost,omitempty"`

	// Host from which to download providers, i.e. a provider registry. See
	// terraform provider registry documentation:
	// https://developer.hashicorp.com/terraform/internals/provider-registry-protocol
	ProviderHost string `yaml:"providerHost,omitempty"`
}

type buildConfig struct {
	MixinConfig

	// AzureUserAgent is the contents of the azure user agent environment variable
	AzureUserAgent string
}

func (m *Mixin) Build(ctx context.Context) error {
	input := BuildInput{
		Config: &m.config, // Apply config directly to the mixin
	}
	err := builder.LoadAction(ctx, m.RuntimeConfig, "", func(contents []byte) (interface{}, error) {
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

	cfg := buildConfig{MixinConfig: *input.Config}
	if !input.Config.UserAgentOptOut {
		cfg.AzureUserAgent = m.userAgent
	}

	return tmpl.Execute(m.Out, cfg)
}
