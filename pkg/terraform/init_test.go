package terraform

import (
	"context"
	"os"
	"testing"

	"get.porter.sh/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestMixin_Init(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	os.Setenv(test.ExpectedCommandEnv, "terraform init")

	h := NewTestMixin(t)

	err := h.Init(context.Background(), nil)

	require.NoError(t, err)
}

func TestMixin_InitBackend(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	os.Setenv(test.ExpectedCommandEnv,
		"terraform init -backend=true -backend-config=donuts=definitely -backend-config=drink=dubonnet -reconfigure")

	h := NewTestMixin(t)

	backendConfig := map[string]interface{}{
		"drink":  "dubonnet",
		"donuts": "definitely",
	}

	err := h.Init(context.Background(), backendConfig)

	require.NoError(t, err)
}
