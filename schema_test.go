// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
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

// author           sigu-399
// author-github    https://github.com/sigu-399
// author-mail      sigu.399@gmail.com
// 
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description      (Unit) Tests for the whole package.	
// 
// created          16-06-2013

package gojsonschema

import (
	"os"
	"testing"
)

// Generic helper function to test validation of a json against a schema
// nbErrorsExpected allows us to make sure failures happen on some cases

func testGeneric(t *testing.T, schemaFilename string, documentFilename string, nbErrorsExpected int) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	schema, err := NewJsonSchemaDocument("file://" + cwd + "/tests/" + schemaFilename)
	if err != nil {
		panic(err.Error())
	}

	jsonToValidate, err := GetFileJson(cwd + "/tests/" + documentFilename)
	if err != nil {
		panic(err.Error())
	}

	validationResult := schema.Validate(jsonToValidate)

	nbErrors := len(validationResult.GetErrorMessages())
	if nbErrors != nbErrorsExpected {
		t.Errorf("Test on ( %s, %s ) failed with %v, %d errors expected ( %d given )", schemaFilename, documentFilename, validationResult.GetErrorMessages(), nbErrorsExpected, nbErrors)
	}

}

func TestSchemaTypes(t *testing.T) {
	testGeneric(t, "schema_types_01.json", "json_types_01_01.json", 0)	// Must be VALID
	testGeneric(t, "schema_types_01.json", "json_types_01_02.json", 6)  // Must FAIL
	testGeneric(t, "schema_types_01.json", "json_types_01_03.json", 6)  // Must FAIL
}

func TestSchemaPresence(t *testing.T) {
	testGeneric(t, "schema_presence_01.json", "json_presence_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_presence_01.json", "json_presence_01_02.json", 1) // Must FAIL
}

func TestSchemaString(t *testing.T) {
	testGeneric(t, "schema_string_01.json", "json_string_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_string_01.json", "json_string_01_02.json", 5) // Must FAIL
}

func TestSchemaNumeric(t *testing.T) {
	testGeneric(t, "schema_numeric_01.json", "json_numeric_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_numeric_01.json", "json_numeric_01_02.json", 11) // Must FAIL
}

func TestSchemaInstance(t *testing.T) {
	testGeneric(t, "schema_instance_01.json", "json_instance_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_instance_01.json", "json_instance_01_02.json", 7) // Must FAIL
}

func TestSchemaArray(t *testing.T) {
	testGeneric(t, "schema_array_01.json", "json_array_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_array_01.json", "json_array_01_02.json", 6) // Must FAIL
}

func TestSchemaObject(t *testing.T) {
	testGeneric(t, "schema_object_01.json", "json_object_01_01.json", 0) // Must be VALID
	testGeneric(t, "schema_object_01.json", "json_object_01_02.json", 1) // Must FAIL
}

func TestSchemaRef(t *testing.T) {
	testGeneric(t, "schema_ref_01.json", "json_ref_01_01.json", 0) // Must be VALID
}
