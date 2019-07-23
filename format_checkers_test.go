package gojsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockFormatChecker struct {
	mock.Mock
}

func (c *mockFormatChecker) IsFormat(input interface{}) bool {
	args := c.Called(input)
	return args.Bool(0)
}

func TestUUIDFormatCheckerIsFormat(t *testing.T) {
	checker := UUIDFormatChecker{}

	assert.True(t, checker.IsFormat("01234567-89ab-cdef-0123-456789abcdef"))
	assert.True(t, checker.IsFormat("f1234567-89ab-cdef-0123-456789abcdef"))

	assert.False(t, checker.IsFormat("not-a-uuid"))
	assert.False(t, checker.IsFormat("g1234567-89ab-cdef-0123-456789abcdef"))
}

func TestURIReferenceFormatCheckerIsFormat(t *testing.T) {
	checker := URIReferenceFormatChecker{}

	assert.True(t, checker.IsFormat("relative"))
	assert.True(t, checker.IsFormat("https://dummyhost.com/dummy-path?dummy-qp-name=dummy-qp-value"))
}
