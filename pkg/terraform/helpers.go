package terraform

import (
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/portercontext"
)

type TestMixin struct {
	*Mixin
	TestContext *portercontext.TestContext
}

// NewTestMixin initializes a terraform mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	c := portercontext.NewTestContext(t)
	m := New()
	m.Context = c.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: c,
	}
}

func sortKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}
