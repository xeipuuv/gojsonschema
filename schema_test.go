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
