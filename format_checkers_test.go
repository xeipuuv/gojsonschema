package gojsonschema

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUUIDFormatCheckerIsFormat(t *testing.T) {
	checker := UUIDFormatChecker{}

	assert.Nil(t, checker.IsFormat("01234567-89ab-cdef-0123-456789abcdef"))
	assert.Nil(t, checker.IsFormat("f1234567-89ab-cdef-0123-456789abcdef"))
	assert.Nil(t, checker.IsFormat("01234567-89AB-CDEF-0123-456789ABCDEF"))
	assert.Nil(t, checker.IsFormat("F1234567-89AB-CDEF-0123-456789ABCDEF"))

	assert.NotNil(t, checker.IsFormat("not-a-uuid"))
	assert.NotNil(t, checker.IsFormat("g1234567-89ab-cdef-0123-456789abcdef"))
}

func TestURIReferenceFormatCheckerIsFormat(t *testing.T) {
	checker := URIReferenceFormatChecker{}

	assert.Nil(t, checker.IsFormat("relative"))
	assert.Nil(t, checker.IsFormat("https://dummyhost.com/dummy-path?dummy-qp-name=dummy-qp-value"))

	err := checker.IsFormat(":")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "missing protocol scheme")
	}

	err = checker.IsFormat("foo\\")
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "contains '\\'")
	}
}

const formatSchema = `{
	"type": "object",
	"properties": {
		"arr":  {"type": "array", "items": {"type": "string"}, "format": "ArrayChecker"},
		"bool": {"type": "boolean", "format": "BoolChecker"},
		"int":  {"format": "IntegerChecker"},
		"name": {"type": "string"},
		"str":  {"type": "string", "format": "StringChecker"}
	},
	"format": "ObjectChecker",
	"required": ["name"]
}`

type arrayChecker struct{}

func (c arrayChecker) IsFormat(input interface{}) error {
	arr, ok := input.([]interface{})
	if !ok {
		return nil
	}
	for _, v := range arr {
		if v == "x" {
			return nil
		}
	}
	return ErrBadFormat
}

type boolChecker struct{}

func (c boolChecker) IsFormat(input interface{}) error {
	b, ok := input.(bool)
	if !ok {
		return nil
	}
	return badFormatUnless(b)
}

type integerChecker struct{}

func (c integerChecker) IsFormat(input interface{}) error {
	number, ok := input.(json.Number)
	if !ok {
		return nil
	}
	f, _ := number.Float64()
	return badFormatUnless(int(f)%2 == 0)
}

type objectChecker struct{}

func (c objectChecker) IsFormat(input interface{}) error {
	obj, ok := input.(map[string]interface{})
	if !ok {
		return nil
	}
	return badFormatUnless(obj["name"] == "x")
}

type stringChecker struct{}

func (c stringChecker) IsFormat(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return nil
	}
	return badFormatUnless(str == "o")
}

func TestCustomFormat(t *testing.T) {
	FormatCheckers.
		Add("ArrayChecker", arrayChecker{}).
		Add("BoolChecker", boolChecker{}).
		Add("IntegerChecker", integerChecker{}).
		Add("ObjectChecker", objectChecker{}).
		Add("StringChecker", stringChecker{})

	sl := NewStringLoader(formatSchema)
	validResult, err := Validate(sl, NewGoLoader(map[string]interface{}{
		"arr":  []string{"x", "y", "z"},
		"bool": true,
		"int":  "2", // format not defined for string
		"name": "x",
		"str":  "o",
	}))
	if err != nil {
		t.Error(err)
	}

	if !validResult.Valid() {
		for _, desc := range validResult.Errors() {
			t.Error(desc)
		}
	}

	invalidResult, err := Validate(sl, NewGoLoader(map[string]interface{}{
		"arr":  []string{"a", "b", "c"},
		"bool": false,
		"int":  1,
		"name": "z",
		"str":  "a",
	}))
	if err != nil {
		t.Error(err)
	}

	assert.Len(t, invalidResult.Errors(), 5)

	FormatCheckers.
		Remove("ArrayChecker").
		Remove("BoolChecker").
		Remove("IntegerChecker").
		Remove("ObjectChecker").
		Remove("StringChecker")
}
