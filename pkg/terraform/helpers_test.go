package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortKeys(t *testing.T) {
	m := map[string]interface{}{
		"delicious": "true",
		"apples":    "green",
		"are":       "needed",
	}

	expected := []string{
		"apples",
		"are",
		"delicious",
	}

	assert.Equal(t, expected, sortKeys(m))
}
