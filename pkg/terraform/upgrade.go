package terraform

// Upgrade runs a terraform apply, just like Install()
func (m *Mixin) Upgrade() error {
	return m.Install()
}
