package gojsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestDateTimeFormatCheckerIsFormat(t *testing.T) {
	checker := DateTimeFormatChecker{}

	assert.False(t, checker.IsFormat("2019-01-01"))
	assert.False(t, checker.IsFormat("2019-01-01 10:00:00"))
	assert.False(t, checker.IsFormat("2019-01-01T10:00:00"))
	assert.True(t, checker.IsFormat("2019-01-01T10:00:00Z"))
}
