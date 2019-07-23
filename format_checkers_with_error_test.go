package gojsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDFormatCheckerIsFormatWithError(t *testing.T) {
	checker := convertToNewChecker(UUIDFormatChecker{})

	isRightFmt, _ := checker.IsFormatWithError("01234567-89ab-cdef-0123-456789abcdef")
	assert.True(t, isRightFmt)

	isRightFmt, _ = checker.IsFormatWithError("f1234567-89ab-cdef-0123-456789abcdef")
	assert.True(t, isRightFmt)

	isRightFmt, err := checker.IsFormatWithError("not-a-uuid")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)

	isRightFmt, err = checker.IsFormatWithError("g1234567-89ab-cdef-0123-456789abcdef")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)
}
