package terraform

import (
	"fmt"
)

const terraformClientVersion = "0.11.11"
const dockerfileLines = `ENV TERRAFORM_VERSION=%s
RUN apt-get update && apt-get install -y wget unzip && \
 wget https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
 unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /usr/bin
COPY . $BUNDLE_DIR
RUN cd %s && terraform init -backend=false
`

func (m *Mixin) Build() error {
	fmt.Fprintf(m.Out, dockerfileLines, terraformClientVersion, m.WorkingDir)
	return nil
}
