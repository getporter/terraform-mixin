package terraform

import "context"

// Upgrade runs a terraform apply, just like Install()
func (m *Mixin) Upgrade(ctx context.Context) error {
	return m.Install(ctx)
}
