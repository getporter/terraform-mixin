package terraform

import (
	"os"
	"testing"

	"github.com/deislabs/porter/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestMixin_Init(t *testing.T) {
	defer os.Unsetenv(test.ExpectedCommandEnv)
	os.Setenv(test.ExpectedCommandEnv, "terraform init")

	h := NewTestMixin(t)

	err := h.Init()

	require.NoError(t, err)
}
