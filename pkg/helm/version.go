package helm

import (
	"fmt"

	"github.com/deislabs/porter-helm/pkg"
)

func (m *Mixin) PrintVersion() {
	fmt.Fprintf(m.Out, "helm mixin %s (%s)\n", pkg.Version, pkg.Commit)
}
