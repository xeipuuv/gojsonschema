package gojsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUUIDFormatCheckerIsFormatWithError(t *testing.T) {
	checker := convertToNewChecker(UUIDFormatChecker{})

	isRightFmt, err := checker.IsFormatWithError("01234567-89ab-cdef-0123-456789abcdef")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)

	isRightFmt, err = checker.IsFormatWithError("f1234567-89ab-cdef-0123-456789abcdef")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)

	isRightFmt, err = checker.IsFormatWithError("not-a-uuid")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)

	isRightFmt, err = checker.IsFormatWithError("g1234567-89ab-cdef-0123-456789abcdef")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)
}

func TestURIReferenceFormatCheckerIsFormatWithError(t *testing.T) {
	checker := convertToNewChecker(URIReferenceFormatChecker{})

	isRightFmt, err := checker.IsFormatWithError("relative")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)

	isRightFmt, err = checker.IsFormatWithError("https://dummyhost.com/dummy-path?dummy-qp-name=dummy-qp-value")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)

	assert.True(t, checker.IsFormat("relative"))
	assert.True(t, checker.IsFormat("https://dummyhost.com/dummy-path?dummy-qp-name=dummy-qp-value"))
}

type mockFormatCheckerWithError struct {
	mock.Mock
}

func (c *mockFormatCheckerWithError) IsFormat(input interface{}) bool {
	args := c.Called(input)
	return args.Bool(0)
}

func (c *mockFormatCheckerWithError) IsFormatWithError(input interface{}) (bool, ResultError) {
	args := c.Called(input)

	b := args.Bool(0)
	rerr, ok := args.Get(1).(ResultError)
	if !ok {
		return b, nil
	}
	return b, rerr
}

func TestGlobalFormatCheckersWithError(t *testing.T) {
	checker := convertToNewChecker(UUIDFormatChecker{})
	fakeFmtTag := "fake"
	FormatCheckers.Add(fakeFmtTag, checker)

	isRightFmt, err := FormatCheckers.IsFormatWithError(fakeFmtTag, "f1234567-89ab-cdef-0123-456789abcdef")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)

	isRightFmt, err = FormatCheckers.IsFormatWithError(fakeFmtTag, "not-a-uuid")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)

	fakeFmtTag2 := "fake2"
	mChecker := new(mockFormatCheckerWithError)
	FormatCheckers.AddCheckerWithError(fakeFmtTag2, mChecker)

	mChecker.On("IsFormatWithError", mock.Anything).Return(true, (ResultError)(nil)).Once()
	isRightFmt, err = FormatCheckers.IsFormatWithError(fakeFmtTag2, "random")
	assert.True(t, isRightFmt)
	assert.Nil(t, err)
	mChecker.AssertExpectations(t)

	customType := "custom"
	customErr := new(ResultErrorFields)
	description := "override format error"
	customErr.SetType(customType)
	customErr.SetDescription(description)
	mChecker.On("IsFormatWithError", mock.Anything).Return(false, customErr).Once()
	isRightFmt, err = FormatCheckers.IsFormatWithError(fakeFmtTag2, "random2")
	assert.False(t, isRightFmt)
	assert.NotNil(t, err)
	assert.Equal(t, customType, err.Type())
	assert.Equal(t, description, err.Description())
	mChecker.AssertExpectations(t)
}
