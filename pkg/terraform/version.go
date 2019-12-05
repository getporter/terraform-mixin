package terraform

import (
	"get.porter.sh/mixin/terraform/pkg"
	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/porter/version"
)

func (m *Mixin) PrintVersion(opts version.Options) error {
	metadata := mixin.Metadata{
		Name: "terraform",
		VersionInfo: mixin.VersionInfo{
			Version: pkg.Version,
			Commit:  pkg.Commit,
			Author:  "DeisLabs",
		},
	}
	return version.PrintVersion(m.Context, opts, metadata)
}
