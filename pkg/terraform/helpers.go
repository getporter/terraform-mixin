package terraform

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"get.porter.sh/porter/pkg/context"
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

func sortKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// Handle the different possible JSON values that the user config could be set to.
func convertValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	t := reflect.TypeOf(value).String()
	switch t {
	case "[]interface {}", "map[string]interface {}":
		bytes, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}
