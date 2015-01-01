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
// description      Result and ResultError implementations.
//
// created          01-01-2015

package gojsonschema

import (
	"fmt"
)

type ResultError struct {
	Context     *jsonContext // Tree like notation of the part that failed the validation. ex (root).a.b ...
	Description string       // A human readable error message
	Value       interface{}  // Value given by the JSON file that is the source of the error
}

func (v ResultError) String() string {

	// as a fallback, the value is displayed go style
	valueString := fmt.Sprintf("%v", v.Value)

	// marshal the go value value to json
	if v.Value == nil {
		valueString = TYPE_NULL
	} else {
		if vs, err := marshalToJsonString(v.Value); err == nil {
			if vs == nil {
				valueString = TYPE_NULL
			} else {
				valueString = *vs
			}
		}
	}

	return fmt.Sprintf(RESULT_ERROR_FORMAT, v.Context, v.Description, valueString)
}

type Result struct {
	errors []ResultError
	// Scores how well the validation matched. Useful in generating
	// better error messages for anyOf and oneOf.
	score int
}

func (v *Result) Valid() bool {
	return len(v.errors) == 0
}

func (v *Result) Errors() []ResultError {
	return v.errors
}

func (v *Result) addError(context *jsonContext, value interface{}, description string) {
	v.errors = append(v.errors, ResultError{Context: context, Value: value, Description: description})
	v.score -= 2 // results in a net -1 when added to the +1 we get at the end of the validation function
}

// Used to copy errors from a sub-schema to the main one
func (v *Result) mergeErrors(otherResult *Result) {
	v.errors = append(v.errors, otherResult.Errors()...)
	v.score += otherResult.score
}

func (v *Result) incrementScore() {
	v.score++
}
