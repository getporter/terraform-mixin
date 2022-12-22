package terraform

import (
	"strconv"
	"strings"

	"get.porter.sh/porter/pkg"
)

const (
	// AzureUserAgentEnvVar is the environment variable used by the azure provider to set
	// the user agent string sent to Azure.
	AzureUserAgentEnvVar = "AZURE_HTTP_USER_AGENT"

	// UserAgentOptOutEnvVar is the name of the environment variable that disables
	// user agent reporting.
	UserAgentOptOutEnvVar = "PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT"
)

// SetUserAgent sets the AZURE_HTTP_USER_AGENT environment variable with
// the full user agent string, which includes both a portion for porter and the
// mixin.
func (m *Mixin) SetUserAgent() {
	// Check if PORTER_TERRAFORM_MIXIN_USER_AGENT_OPT_OUT=true, which disables editing the user agent string
	if optOut, _ := strconv.ParseBool(m.Getenv(UserAgentOptOutEnvVar)); optOut {
		return
	}

	// Check if we have already set the user agent
	if m.userAgent != "" {
		return
	}

	porterUserAgent := pkg.UserAgent()
	mixinUserAgent := m.GetMixinUserAgent()
	userAgent := []string{porterUserAgent, mixinUserAgent}
	// Append porter and the mixin's version to the user agent string. Some clouds and
	// environments will have set the environment variable already and we don't want
	// to clobber it.
	value := strings.Join(userAgent, " ")
	if agentStr, ok := m.LookupEnv(AzureUserAgentEnvVar); ok {

		// Check if we have already set the user agent
		if strings.Contains(agentStr, value) {
			value = agentStr
		} else {
			userAgent = append(userAgent, agentStr)
			value = strings.Join(userAgent, " ")
		}
	}

	m.userAgent = value

	// Set the environment variable so that when we call the
	// azure provider, it's automatically passed too.
	m.Setenv(AzureUserAgentEnvVar, m.userAgent)
}

// GetMixinUserAgent returns the portion of the user agent string for the mixin.
func (m *Mixin) GetMixinUserAgent() string {
	v := m.Version()
	return "getporter/" + v.Name + "/" + v.Version
}
