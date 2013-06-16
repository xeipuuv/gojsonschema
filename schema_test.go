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

func testGeneric(t *testing.T, schemaFilename string, documentFilename string, nbErrorsExpected int) {

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	schema, err := NewJsonSchemaDocument("file://" + cwd + "/examples/" + schemaFilename)
	if err != nil {
		panic(err.Error())
	}

	jsonToValidate, err := GetFileJson(cwd + "/examples/" + documentFilename)
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

	testGeneric(t, "schema_types_01.json", "json_types_01_01.json", 0)
	testGeneric(t, "schema_types_01.json", "json_types_01_02.json", 6)
	testGeneric(t, "schema_types_01.json", "json_types_01_03.json", 6)

}
