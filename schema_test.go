// Copyright 2015 xeipuuv ( https://github.com/xeipuuv )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// author           xeipuuv
// author-github    https://github.com/xeipuuv
// author-mail      xeipuuv@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description      (Unit) Tests for schema validation.
//
// created          16-06-2013

package gojsonschema

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const displayErrorMessages = false

const circularReference = `{
	"type": "object",
	"properties": {
		"games": {
			"type": "array",
			"items": {
				"$ref": "#/definitions/game"
			}
		}
	},
	"definitions": {
		"game": {
			"type": "object",
			"properties": {
				"winner": {
					"$ref": "#/definitions/player"
				},
				"loser": {
					"$ref": "#/definitions/player"
				}
			}
		},
		"player": {
			"type": "object",
			"properties": {
				"user": {
					"$ref": "#/definitions/user"
				},
				"game": {
					"$ref": "#/definitions/game"
				}
			}
		},
		"user": {
			"type": "object",
			"properties": {
				"fullName": {
					"type": "string"
				}
			}
		}
	}
}`

func TestCircularReference(t *testing.T) {
	loader := NewStringLoader(circularReference)
	// call the target function
	_, err := NewSchema(loader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
}

// From http://json-schema.org/examples.html
const simpleSchema = `{
  "title": "Example Schema",
  "type": "object",
  "properties": {
    "firstName": {
      "type": "string"
    },
    "lastName": {
      "type": "string"
    },
    "age": {
      "description": "Age in years",
      "type": "integer",
      "minimum": 0
    }
  },
  "required": ["firstName", "lastName"]
}`

func TestLoaders(t *testing.T) {
	// setup reader loader
	reader := bytes.NewBufferString(simpleSchema)
	readerLoader, wrappedReader := NewReaderLoader(reader)

	// drain reader
	by, err := ioutil.ReadAll(wrappedReader)
	assert.Nil(t, err)
	assert.Equal(t, simpleSchema, string(by))

	// setup writer loaders
	writer := &bytes.Buffer{}
	writerLoader, wrappedWriter := NewWriterLoader(writer)

	// fill writer
	n, err := io.WriteString(wrappedWriter, simpleSchema)
	assert.Nil(t, err)
	assert.Equal(t, n, len(simpleSchema))

	loaders := []JSONLoader{
		NewStringLoader(simpleSchema),
		readerLoader,
		writerLoader,
	}

	for _, l := range loaders {
		_, err := NewSchema(l)
		assert.Nil(t, err, "loader: %T", l)
	}
}

const invalidPattern = `{
  "title": "Example Pattern",
  "type": "object",
  "properties": {
    "invalid": {
      "type": "string",
      "pattern": 99999
    }
  }
}`

func TestLoadersWithInvalidPattern(t *testing.T) {
	// setup reader loader
	reader := bytes.NewBufferString(invalidPattern)
	readerLoader, wrappedReader := NewReaderLoader(reader)

	// drain reader
	by, err := ioutil.ReadAll(wrappedReader)
	assert.Nil(t, err)
	assert.Equal(t, invalidPattern, string(by))

	// setup writer loaders
	writer := &bytes.Buffer{}
	writerLoader, wrappedWriter := NewWriterLoader(writer)

	// fill writer
	n, err := io.WriteString(wrappedWriter, invalidPattern)
	assert.Nil(t, err)
	assert.Equal(t, n, len(invalidPattern))

	loaders := []JSONLoader{
		NewStringLoader(invalidPattern),
		readerLoader,
		writerLoader,
	}

	for _, l := range loaders {
		_, err := NewSchema(l)
		assert.NotNil(t, err, "expected error loading invalid pattern: %T", l)
	}
}

const refPropertySchema = `{
	"$id" : "http://localhost/schema.json",
	"properties" : {
		"$id" : {
			"$id": "http://localhost/foo.json"
		},
		"$ref" : {
			"const": {
				"$ref" : "hello.world"
			}
		},
		"const" : {
			"$ref" : "#/definitions/$ref"
		}
	},
	"definitions" : {
		"$ref" : {
			"const": {
				"$ref" : "hello.world"
			}
		}
	},
	"dependencies" : {
		"$ref" : [ "const" ],
		"const" : [ "$ref" ]
	}
}`

func TestRefProperty(t *testing.T) {
	schemaLoader := NewStringLoader(refPropertySchema)
	documentLoader := NewStringLoader(`{
		"$ref" : { "$ref" : "hello.world" },
		"const" : { "$ref" : "hello.world" }
		}`)
	// call the target function
	s, err := NewSchema(schemaLoader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
	result, err := s.Validate(documentLoader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}
	if !result.Valid() {
		for _, err := range result.Errors() {
			fmt.Println(err.String())
		}
		t.Errorf("Got invalid validation result.")
	}
}

func TestFragmentLoader(t *testing.T) {
	wd, err := os.Getwd()

	if err != nil {
		panic(err.Error())
	}

	fileName := filepath.Join(wd, "testdata", "extra", "fragment_schema.json")

	schemaLoader := NewReferenceLoader("file://" + filepath.ToSlash(fileName) + "#/definitions/x")
	schema, err := NewSchema(schemaLoader)

	if err != nil {
		t.Errorf("Encountered error while loading schema: %s", err.Error())
	}

	validDocument := NewStringLoader(`5`)
	invalidDocument := NewStringLoader(`"a"`)

	result, err := schema.Validate(validDocument)

	if assert.Nil(t, err, "Unexpected error while validating document: %T", err) {
		if !result.Valid() {
			t.Errorf("Got invalid validation result.")
		}
	}

	result, err = schema.Validate(invalidDocument)

	if assert.Nil(t, err, "Unexpected error while validating document: %T", err) {
		if len(result.Errors()) != 1 || result.Errors()[0].Type() != "invalid_type" {
			t.Errorf("Got invalid validation result.")
		}
	}
}

func TestFileWithSpace(t *testing.T) {
	wd, err := os.Getwd()

	if err != nil {
		panic(err.Error())
	}

	fileName := filepath.Join(wd, "testdata", "extra", "file with space.json")
	loader := NewReferenceLoader("file://" + filepath.ToSlash(fileName))

	json, err := loader.LoadJSON()

	assert.Nil(t, err, "Unexpected error when trying to load a filepath containing a space")
	assert.Equal(t, map[string]interface{}{"foo": true}, json, "Contents of the file do not match")
}

func TestAdditionalPropertiesErrorMessage(t *testing.T) {
	schema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "Device": {
      "type": "object",
      "additionalProperties": {
        "type": "string"
      }
    }
  }
}`
	text := `{
		"Device":{
			"Color" : true
		}
	}`
	loader := NewBytesLoader([]byte(schema))
	result, err := Validate(loader, NewBytesLoader([]byte(text)))
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Errors()) != 1 {
		t.Fatal("Expected 1 error but got", len(result.Errors()))
	}

	expected := "Device.Color: Invalid type. Expected: string, given: boolean"
	actual := result.Errors()[0].String()
	if actual != expected {
		t.Fatalf("Expected '%s' but got '%s'", expected, actual)
	}
}

// Inspired by http://json-schema.org/latest/json-schema-core.html#rfc.section.8.2.3
const locationIndependentSchema = `{
  "definitions": {
    "A": {
      "$id": "#foo"
    },
    "B": {
      "$id": "http://example.com/other.json",
      "definitions": {
        "X": {
          "$id": "#bar",
          "allOf": [false]
        },
        "Y": {
          "$id": "t/inner.json"
        }
      }
    },
    "C": {
			"$id" : "#frag",
      "$ref": "http://example.com/other.json#bar"
    }
  },
  "$ref": "#frag"
}`

func TestLocationIndependentIdentifier(t *testing.T) {
	schemaLoader := NewStringLoader(locationIndependentSchema)
	documentLoader := NewStringLoader(`{}`)

	s, err := NewSchema(schemaLoader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}

	result, err := s.Validate(documentLoader)
	if err != nil {
		t.Errorf("Got error: %s", err.Error())
	}

	if len(result.Errors()) != 2 || result.Errors()[0].Type() != "false" || result.Errors()[1].Type() != "number_all_of" {
		t.Errorf("Got invalid validation result.")
	}
}

const incorrectRefSchema = `{
  "$ref" : "#/fail"
}`

const incorrectRefSchema2 = `{
	"$ref" : 123
}`

func TestIncorrectRef(t *testing.T) {

	schemaLoader := NewStringLoader(incorrectRefSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t, s)
	assert.Equal(t, "Object has no key 'fail'", err.Error())

	schemaLoader2 := NewStringLoader(incorrectRefSchema2)
	s, err = NewSchema(schemaLoader2)
	assert.Nil(t, s)
	assert.Equal(t, "Invalid type. Expected: string, given: $ref", err.Error())
}

func TestIncorrectId(t *testing.T) {

	const incorrectIdSchema = `{
		"schema": "http://json-schema.org/draft-07/schema#",
		"$id": 123
	}`

	schemaLoader := NewStringLoader(incorrectIdSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t, "Invalid type. Expected: string, given: $id", err.Error())
}

func TestIncorrectDefinitions(t *testing.T) {

	const incorrectDefinitionsSchema1 = `{
		"schema": "http://json-schema.org/draft-04/schema#",
		"definitions" : 123
	}`
	const incorrectDefinitionsSchema2 = `{
		"schema": "http://json-schema.org/draft-04/schema#",
		"definitions": {"foo": 1}
	}`

	schemaLoader1 := NewStringLoader(incorrectDefinitionsSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t, "Invalid type. Expected: array of schemas, given: definitions", err.Error())

	schemaLoader2 := NewStringLoader(incorrectDefinitionsSchema2)
	s, err = NewSchema(schemaLoader2)
	assert.Nil(t,s)
	assert.Equal(t, "Invalid type. Expected: array of schemas, given: definitions", err.Error())
}

func TestIncorrectTitle(t *testing.T) {

	const incorrectTitleSchema = `{
		"schema": "http://json-schema.org/draft-04/schema#",
		"title": 123
	}`

	schemaLoader := NewStringLoader(incorrectTitleSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: string, given: title", err.Error())
}

func TestIncorrectPatternPorperties(t *testing.T) {
	
	const incorrectPatternPropertiesSchema = `{
		"schema": "http://json-schema.org/draft-04/schema#",
		"patternProperties": 123
	}`

	schemaLoader := NewStringLoader(incorrectPatternPropertiesSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: valid schema, given: patternProperties", err.Error())
}

func TestIncorrectDescription(t *testing.T) {

	const incorrectDescriptionSchema = `{
		"schema": "http://json-schema.org/draft-04/schema#",
		"description": 123
	}`

	schemaLoader := NewStringLoader(incorrectDescriptionSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: string, given: description", err.Error())
}

func TestIncorrectType(t *testing.T) {

	const incorrectTypeSchema = `{"type": 1}`

	schemaLoader := NewStringLoader(incorrectTypeSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: string/array of strings, given: type", err.Error())
}

func TestIncorrectAdditionalItems(t *testing.T) {

	const incorrectAdditionalItemsSchema = `{"additionalItems": 123}`

	schemaLoader := NewStringLoader(incorrectAdditionalItemsSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: boolean/valid schema, given: additionalItems", err.Error())
}

func TestIncorrectAdditionalProperties(t *testing.T) {

	const incorrectAdditionalPropertiesSchema = `{"additionalProperties": 123}`

	schemaLoader := NewStringLoader(incorrectAdditionalPropertiesSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: boolean/valid schema, given: additionalProperties", err.Error())
}

func TestIncorrectMultipleOf(t *testing.T) {

	const incorrectMultipleOfSchema1 = `{"multipleOf": ""}`
	const incorrectMultipleOfSchema2 = `{"multipleOf": 0}`

	schemaLoader1 := NewStringLoader(incorrectMultipleOfSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: number, given: multipleOf", err.Error())

	schemaLoader2 := NewStringLoader(incorrectMultipleOfSchema2)
	s, err = NewSchema(schemaLoader2)

	assert.Nil(t,s)
	assert.Equal(t,"multipleOf must be strictly greater than 0", err.Error())
}

func TestIncorrectMinimum(t *testing.T) {

	const incorrectMinimumSchema = `{"minimum": ""}`

	schemaLoader := NewStringLoader(incorrectMinimumSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"minimum must be of a number", err.Error())
}

func TestIncorrectExclusiveMinimum(t *testing.T) {

	const incorrectExclusiveMinimumSchema1 = `{"exclusiveMinimum": true}`
	const incorrectExclusiveMinimumSchema2 = `{"exclusiveMinimum": ""}`

	schemaLoader1 := NewStringLoader(incorrectExclusiveMinimumSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t,"exclusiveMinimum cannot be used without minimum", err.Error())

	schemaLoader2 := NewStringLoader(incorrectExclusiveMinimumSchema2)
	s, err = NewSchema(schemaLoader2)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: boolean/number, given: exclusiveMinimum", err.Error())
}

func TestIncorrectMaximum(t *testing.T) {

	const incorrectMaximumSchema = `{"maximum": ""}`

	schemaLoader := NewStringLoader(incorrectMaximumSchema)
	s, err := NewSchema(schemaLoader)

	assert.Nil(t,s)
	assert.Equal(t,"maximum must be of a number", err.Error())
}

func TestIncorrectExclusiveMaximum(t *testing.T) {

	const incorrectExclusiveMaximumSchema1 = `{"exclusiveMaximum": ""}`
	const incorrectExclusiveMaximumSchema2 = `{"exclusiveMaximum": true}`

	schemaLoader1 := NewStringLoader(incorrectExclusiveMaximumSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t,"Invalid type. Expected: boolean/number, given: exclusiveMaximum", err.Error())

	schemaLoader2 := NewStringLoader(incorrectExclusiveMaximumSchema2)
	s, err = NewSchema(schemaLoader2)

	assert.Nil(t,s)
	assert.Equal(t,"exclusiveMaximum cannot be used without maximum", err.Error())
}

func TestIncorrectMinLength(t *testing.T) {

	const incorrectMinLengthSchema1 = `{"minLength": ""}`
	const incorrectMinLengthSchema2 = `{"minLength": -1}`

	schemaLoader1 := NewStringLoader(incorrectMinLengthSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t,"minLength must be of an integer", err.Error())

	schemaLoader2 := NewStringLoader(incorrectMinLengthSchema2)
	s, err = NewSchema(schemaLoader2)

	assert.Nil(t,s)
	assert.Equal(t,"minLength must be greater than or equal to 0", err.Error())
}

func TestIncorrectMaxLength(t *testing.T) {

	const incorrectMaxLengthSchema1 = `{"maxLength": ""}`
	const incorrectMaxLengthSchema2 = `{"maxLength": -1}`

	schemaLoader1 := NewStringLoader(incorrectMaxLengthSchema1)
	s, err := NewSchema(schemaLoader1)

	assert.Nil(t,s)
	assert.Equal(t,"maxLength must be of an integer", err.Error())

	schemaLoader2 := NewStringLoader(incorrectMaxLengthSchema2)
	s, err = NewSchema(schemaLoader2)

	assert.Nil(t,s)
	assert.Equal(t,"maxLength must be greater than or equal to 0", err.Error())
}
