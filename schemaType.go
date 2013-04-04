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

// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Helper structure to handle schema types, and the combination of them.			
// 
// created      	28-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"strings"
)

type jsonSchemaType struct {
	types []string
}

func (t *jsonSchemaType) Add(etype string) error {

	if !isStringInSlice(JSON_TYPES, etype) {
		return errors.New(fmt.Sprintf("%s is not a valid type", etype))
	}

	if t.HasType(etype) {
		return errors.New(fmt.Sprintf("%s type is duplicated", etype))
	}

	t.types = append(t.types, etype)

	return nil
}

func (t *jsonSchemaType) HasType(etype string) bool {

	for _, v := range t.types {
		if v == etype {
			return true
		}
	}

	return false
}

func (t *jsonSchemaType) String() string {
	return strings.Join(t.types, ",")
}
