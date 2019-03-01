package terraform

import (
	"fmt"
)

const terraformClientVersion = "v2.12.3"
const dockerfileLines = `RUN apt-get update && \
 apt-get install -y curl && \
 curl -o terraform.tgz https://storage.googleapis.com/kubernetes-terraform/terraform-%s-linux-amd64.tar.gz && \
 tar -xzf terraform.tgz && \
 mv linux-amd64/terraform /usr/local/bin && \
 rm terraform.tgz
RUN terraform init --client-only`

func (m *Mixin) Build() error {
	fmt.Fprintf(m.Out, dockerfileLines, terraformClientVersion)
	return nil
}
