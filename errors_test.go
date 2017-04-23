// Copyright 2016 Ravelin ( https://github.com/unravelin )
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

// author           sweeney
// author-github    https://github.com/sweeney
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description     (Unit) Tests for errors
//
// created          08-04-2016

package gojsonschema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUsefulErrorDescription(t *testing.T) {

	//	assert.Equal(t, "ddd", "ddde")

	noField := usefulErrorDescription(
		"",
		"%dog%<>%cat%",
		ErrorDetails{
			"dog": "cat",
			"cat": "dog",
		},
	)

	someField := usefulErrorDescription(
		"field",
		"%dog%<>%cat%",
		ErrorDetails{
			"dog": "cat",
			"cat": "dog",
		},
	)

	rootField := usefulErrorDescription(
		STRING_CONTEXT_ROOT,
		"%dog%<>%cat%",
		ErrorDetails{
			"dog": "cat",
			"cat": "dog",
		},
	)

	rootishField := usefulErrorDescription(
		"(root)",
		"%dog%<>%cat%",
		ErrorDetails{
			"dog": "cat",
			"cat": "dog",
		},
	)

	assert.Equal(t, "cat<>dog", noField)
	assert.Equal(t, "field: cat<>dog", someField)
	assert.Equal(t, "cat<>dog", rootField)
	assert.Equal(t, "cat<>dog", rootishField)

}
