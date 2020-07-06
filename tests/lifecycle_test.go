// +build integration

package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"get.porter.sh/porter/pkg/porter"
)

// Entire lifecycle test of a sample terraform bundle
func TestLifecycle(t *testing.T) {
	p := porter.NewTestPorter(t)

	p.SetupIntegrationTest()
	defer p.CleanupIntegrationTest()
	p.Debug = false

	// Install the bundle that has dependencies
	p.CopyDirectory(filepath.Join(p.TestDir, "../build/testdata/bundles/terraform"), ".", false)

	installOpts := porter.InstallOptions{
		BundleLifecycleOpts: porter.BundleLifecycleOpts{
			SharedOptions: porter.SharedOptions{
				Params: []string{
					{
						"file_contents='foo!'",
					},
				},
			},
		},
	}
	err := installOpts.Validate([]string{}, p.Context)
	require.NoError(t, err)

	err = p.InstallBundle(installOpts)
	require.NoError(t, err)
}
