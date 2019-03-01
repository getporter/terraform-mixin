package terraform

import (
	"fmt"

	"github.com/deislabs/porter-terraform/pkg"
)

func (m *Mixin) PrintVersion() {
	fmt.Fprintf(m.Out, "terraform mixin %s (%s)\n", pkg.Version, pkg.Commit)
}
