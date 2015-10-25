// Author: Bodie Solomon
//         bodie@synapsegarden.net
//         github.com/binary132
//
//         2015-02-16

package gojsonschema_test

import (
	"errors"
	"fmt"
	"log"
	"testing"

	gjs "github.com/eduardonunesp/gojsonschema"
	"github.com/stretchr/testify/assert"
)

func M(in ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if len(in)%2 != 0 {
		log.Fatal("map construction M must have one value for each key")
	}

	for i := 0; i < len(in); i += 2 {
		// Keys must be strings
		k := in[i]
		v := in[i+1]
		sK := k.(string)
		result[sK] = v
	}

	return result
}

// Get a *Schema given a properties map.
func schemaFromProperties(t *testing.T, properties map[string]interface{}) *gjs.Schema {
	schemaMap := M("type", "object", "properties", properties)
	loader := gjs.NewGoLoader(schemaMap)
	schema, err := gjs.NewSchema(loader)
	assert.Nil(t, err)
	return schema
}

func testGetDocProperties(schema *gjs.Schema) (doc map[string]interface{}, testError error) {
	// Capture expected panics
	defer func() {
		if r := recover(); r != nil {
			var msg string
			switch typed := r.(type) {
			case error:
				msg = fmt.Sprintf("panic: %s", typed.Error())
			default:
				msg = fmt.Sprintf("unexpected panic: %#v", typed)
			}
			testError = errors.New(msg)
		}
	}()
	schemaDoc := schema.GetDocProperties()
	return schemaDoc, nil
}

func TestGetDocProperties(t *testing.T) {
	tests := []struct {
		should         string
		usingSchema    interface{}
		expectedErr    string
		expectedResult interface{}
	}{{
		should:      "panic for nil pool",
		expectedErr: "runtime error: invalid memory address or nil pointer dereference",
	}, {
		should:      "panic for non-map document",
		usingSchema: "not-a-map",
		expectedErr: "interface conversion: interface {} is string, not map[string]interface {}",
	}, {
		should:      "panic for non schema document",
		usingSchema: M("foo", "bar"),
		expectedErr: "interface conversion: interface {} is nil, not map[string]interface {}",
	}, {
		should:         "work for an OK document",
		usingSchema:    M("properties", M("foo", "bar")),
		expectedResult: M("foo", "bar"),
	}}

	for i, test := range tests {
		fmt.Printf("TestGetDocProperties test (%d) | should %s\n", i, test.should)
		schema := gjs.MakeTestingSchema(test.usingSchema)
		doc, err := testGetDocProperties(schema)
		if test.expectedErr != "" {
			assert.EqualError(t, err, "panic: "+test.expectedErr)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedResult, doc)
	}

	println()
}

func TestIterateAndInsert(t *testing.T) {
	tests := []struct {
		should          string
		usingProperties map[string]interface{}
		into            map[string]interface{}
		expectedResult  map[string]interface{}
	}{{
		should:          "insert value as expected for a simple default",
		usingProperties: M("num", M("default", 5, "type", "integer")),
		into:            M(),
		expectedResult:  M("num", 5),
	}, {
		should:          "not overwrite existing values",
		usingProperties: M("num", M("default", 5, "type", "integer")),
		into:            M("num", 8),
		expectedResult:  M("num", 8),
	}, {
		should:          "create a simple map with a default inner value",
		usingProperties: M("num", M("properties", M("dum", M("default", 5)))),
		into:            M(),
		expectedResult:  M("num", M("dum", 5)),
	}, {
		should:          "non-destructively insert a value into an inner map",
		usingProperties: M("num", M("properties", M("dum", M("default", 5)))),
		into:            M("num", M("gum", 8)),
		expectedResult:  M("num", M("gum", 8, "dum", 5)),
	}, {
		should:          "non-destructively insert a value into an inner map",
		usingProperties: M("num", M("properties", M("dum", M("default", 5)))),
		into:            M("num", M("gum", 8), "foo", "bar"),
		expectedResult:  M("num", M("gum", 8, "dum", 5), "foo", "bar"),
	}, {
		should: "non-destructively insert a complex default into an inner map",
		usingProperties: M(
			"num", M("properties",
				M("dum", M("default", 5),
					"foo", M("default", M("bar", "baz")))),
			"foo", M("properties",
				M("bar", M("default", "baz")))),
		into: M("num", M("gum", 8), "foo", M()),
		expectedResult: M(
			"num", M("gum", 8, "dum", 5, "foo", M("bar", "baz")),
			"foo", M("bar", "baz")),
	}, {
		should: "not insert a value if there is no default",
		usingProperties: M(
			"num", M("properties",
				M("dum", M("type", "string",
					"description", "something dum"),
					"foo", M("default", M("bar", "baz")))),
			"foo", M("properties",
				M("bar", M("baz", M("woz", M("type", "integer")))))),
		into:           M(),
		expectedResult: M("num", M("foo", M("bar", "baz"))),
	}, {
		should:          "ignore bad values",
		usingProperties: M("foo", M("properties", M("bar", M("default", 5)))),
		into:            M("foo", 5),
		expectedResult:  M("foo", 5),
	}}

	for i, test := range tests {
		fmt.Printf("TestIterateAndInsert test (%d) | should %s\n", i, test.should)
		gjs.IterateAndInsert(test.into, test.usingProperties)
		assert.Equal(t, test.expectedResult, test.into)
	}

	println()
}

func TestInsertDefaults(t *testing.T) {
	tests := []struct {
		should          string
		usingProperties map[string]interface{}
		into            map[string]interface{}
		expectedResult  map[string]interface{}
		expectedError   string
	}{{
		should:          "have no problem with empty properties",
		usingProperties: M(),
		expectedResult:  M(),
	}, {
		should:          "handle panics from bad schemas gracefully",
		usingProperties: M("foo", "bar"),
		expectedError:   "interface conversion: interface {} is string, not map[string]interface {}",
	}, {
		should:          "return empty map if given nil target",
		into:            nil,
		usingProperties: M(),
		expectedResult:  M(),
	}, {
		should:          "make map if given nil target",
		into:            nil,
		usingProperties: M("foo", M("default", 5)),
		expectedResult:  M("foo", 5),
	}, {
		should:          "work for simple schemas",
		usingProperties: M("foo", M("default", 5)),
		into:            M("bar", 4),
		expectedResult:  M("bar", 4, "foo", 5),
	}, {
		should:          "not overwrite values",
		usingProperties: M("foo", M("default", 5)),
		into:            M("foo", 4),
		expectedResult:  M("foo", 4),
	}, {
		should: "work for more complex schemas",
		usingProperties: M(
			"num", M("properties", M(
				"dum", M("type", "string", "description", "dum"),
				"foo", M("default", M("bar", "baz")))),
			"foo", M("properties", M(
				"bar", M("properties", M(
					"baz", M("properties", M(
						"woz", M("type", "integer")))))))),
		into:           M(),
		expectedResult: M("num", M("foo", M("bar", "baz"))),
	}}

	for i, test := range tests {
		fmt.Printf("TestDefaultObjects test (%d) | should %s\n", i, test.should)
		schemaDoc := M("properties", test.usingProperties)
		schema := gjs.MakeTestingSchema(schemaDoc)
		result, err := schema.InsertDefaults(test.into)
		if test.expectedError != "" {
			assert.EqualError(t, err, "schema error caused a panic: "+test.expectedError)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedResult, result)
	}

	println()
}
