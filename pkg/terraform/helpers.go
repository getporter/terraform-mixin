package terraform

import (
	"testing"

	"github.com/deislabs/porter/pkg/context"
)

type TestMixin struct {
	*Mixin
	TestContext *context.TestContext
}

// NewTestMixin initializes a terraform mixin, with the output buffered, and an in-memory file system.
func NewTestMixin(t *testing.T) *TestMixin {
	c := context.NewTestContext(t)
	m := New()
	m.Context = c.Context
	return &TestMixin{
		Mixin:       m,
		TestContext: c,
	}
}
