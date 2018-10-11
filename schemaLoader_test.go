// Copyright 2018 johandorland ( https://github.com/johandorland )
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

package gojsonschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaLoaderWithReferenceToAddedSchema(t *testing.T) {
	ps := NewSchemaLoader()
	err := ps.AddSchemas(NewStringLoader(`{
		"$id" : "http://localhost:1234/test1.json",
		"type" : "integer"
		}`))

	assert.Nil(t, err)
	schema, err := ps.Compile(NewReferenceLoader("http://localhost:1234/test1.json"))
	assert.Nil(t, err)
	result, err := schema.Validate(NewStringLoader(`"hello"`))
	assert.Nil(t, err)
	if len(result.Errors()) != 1 || result.Errors()[0].Type() != "invalid_type" {
		t.Errorf("Expected invalid type erorr, instead got %v", result.Errors())
	}
}

func TestCrossReference(t *testing.T) {
	schema1 := NewStringLoader(`{
		"$ref" : "http://localhost:1234/test3.json",
		"definitions" : {
			"foo" : {
				"type" : "integer"
			}
		}
	}`)
	schema2 := NewStringLoader(`{
		"$ref" : "http://localhost:1234/test2.json#/definitions/foo"
	}`)

	ps := NewSchemaLoader()
	err := ps.AddSchema("http://localhost:1234/test2.json", schema1)
	assert.Nil(t, err)
	err = ps.AddSchema("http://localhost:1234/test3.json", schema2)
	assert.Nil(t, err)
	schema, err := ps.Compile(NewStringLoader(`{"$ref" : "http://localhost:1234/test2.json"}`))
	assert.Nil(t, err)
	result, err := schema.Validate(NewStringLoader(`"hello"`))
	assert.Nil(t, err)
	if len(result.Errors()) != 1 || result.Errors()[0].Type() != "invalid_type" {
		t.Errorf("Expected invalid type erorr, instead got %v", result.Errors())
	}
}

// Multiple schemas identifying under the same $id should throw an error
func TestDoubleIDReference(t *testing.T) {
	ps := NewSchemaLoader()
	err := ps.AddSchema("http://localhost:1234/test4.json", NewStringLoader("{}"))
	assert.Nil(t, err)
	err = ps.AddSchemas(NewStringLoader(`{ "$id" : "http://localhost:1234/test4.json"}`))
	assert.NotNil(t, err)
}
