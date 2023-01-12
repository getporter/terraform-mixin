package terraform

import (
	"os"
	"testing"

	"get.porter.sh/mixin/terraform/pkg"
	"get.porter.sh/porter/pkg/runtime"
	"github.com/stretchr/testify/require"
)

func TestMixinSetsUserAgentEnvVar(t *testing.T) {
	// CI sets this value and we need to clear it out to make the test reproducible
	os.Unsetenv(AzureUserAgentEnvVar)

	t.Run("sets env var", func(t *testing.T) {
		pkg.Commit = "abc123"
		pkg.Version = "v1.2.3"
		m := New()
		expected := "getporter/porter getporter/terraform/" + pkg.Version
		require.Equal(t, expected, m.Getenv(AzureUserAgentEnvVar))
		require.Equal(t, expected, m.userAgent, "validate we remember the user agent string for later")
	})
	t.Run("edits env var", func(t *testing.T) {
		os.Unsetenv(AzureUserAgentEnvVar)
		// Validate that if the user customizations of the env var are preserved
		pkg.Commit = "abc123"
		pkg.Version = "v1.2.3"
		cfg := runtime.NewConfig()
		customUserAgent := "mycustom/v1.2.3"
		cfg.Setenv(AzureUserAgentEnvVar, customUserAgent)
		m := NewFor(cfg)
		expected := "getporter/porter getporter/terraform/v1.2.3 mycustom/v1.2.3"
		require.Equal(t, expected, m.Getenv(AzureUserAgentEnvVar))
		require.Equal(t, expected, m.userAgent, "validate we remember the user agent string for later")
	})

	t.Run("env var already set", func(t *testing.T) {
		// Validate that calling multiple times doesn't make a messed up env var
		os.Unsetenv(AzureUserAgentEnvVar)
		pkg.Commit = "abc123"
		pkg.Version = "v1.2.3"
		cfg := runtime.NewConfig()
		customUserAgent := "getporter/porter getporter/terraform/v1.2.3"
		cfg.Setenv(AzureUserAgentEnvVar, customUserAgent)
		m := New()
		m.SetUserAgent()
		expected := "getporter/porter getporter/terraform/v1.2.3"
		require.Equal(t, expected, m.Getenv(AzureUserAgentEnvVar))
		require.Equal(t, expected, m.userAgent, "validate we remember the user agent string for later")
	})
	t.Run("call multiple times", func(t *testing.T) {
		// Validate that calling multiple times doesn't make a messed up env var
		os.Unsetenv(AzureUserAgentEnvVar)
		pkg.Commit = "abc123"
		pkg.Version = "v1.2.3"
		m := New()
		m.SetUserAgent()
		m.SetUserAgent()
		expected := "getporter/porter getporter/terraform/v1.2.3"
		require.Equal(t, expected, m.Getenv(AzureUserAgentEnvVar))
		require.Equal(t, expected, m.userAgent, "validate we remember the user agent string for later")
	})
}

func TestMixinSetsUserAgentEnvVar_OptOut(t *testing.T) {
	// CI sets this value and we need to clear it out to make the test reproducible
	os.Unsetenv(AzureUserAgentEnvVar)

	t.Run("opt-out", func(t *testing.T) {
		// Validate that at runtime when we are calling the az cli, that if the bundle author opted-out, we don't set the user agent string
		cfg := runtime.NewConfig()
		cfg.Setenv(UserAgentOptOutEnvVar, "true")
		m := NewFor(cfg)
		_, hasEnv := m.LookupEnv(AzureUserAgentEnvVar)
		require.False(t, hasEnv, "expected the opt out to skip setting the AZURE_HTTP_USER_AGENT environment variable")
	})
	t.Run("opt-out preserves original value", func(t *testing.T) {
		// Validate that at runtime when we are calling the az cli, that if the bundle author opted-out, we don't set the user agent string
		cfg := runtime.NewConfig()
		cfg.Setenv(UserAgentOptOutEnvVar, "true")
		customUserAgent := "mycustom/v1.2.3"
		cfg.Setenv(AzureUserAgentEnvVar, customUserAgent)
		m := NewFor(cfg)
		require.Equal(t, customUserAgent, m.Getenv(AzureUserAgentEnvVar), "expected opting out to not prevent the user from setting a custom user agent")
		require.Empty(t, m.userAgent, "validate we remember that we opted out")
	})
}
