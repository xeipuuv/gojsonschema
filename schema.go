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
// description      Defines the structure of a schema. 
//                  A schema can have sub-schemas.
// 
// created          27-02-2013

package gojsonschema

import (
	"errors"
	"github.com/sigu-399/gojsonreference"
	"regexp"
)

type jsonSchema struct {

	// basic schema meta properties
	id          *string
	title       *string
	description *string
	
	types       jsonSchemaType

	ref *gojsonreference.JsonReference
	schema *gojsonreference.JsonReference

	// hierarchy 
	parent *jsonSchema

	definitionsChildren []*jsonSchema
	itemsChild          *jsonSchema
	propertiesChildren  []*jsonSchema
	
	property string

	// validation : number / integer
	multipleOf       *float64
	maximum          *float64
	exclusiveMaximum bool
	minimum          *float64
	exclusiveMinimum bool

	// validation : string
	minLength *int
	maxLength *int
	pattern   *regexp.Regexp

	// validation : object
	minProperties *int
	maxProperties *int
	required []string
	
	dependencies map[string][]string

	// validation : array
	minItems    *int
	maxItems    *int
	uniqueItems bool

	// validation : all
	enum []string

	// validation : schema
	oneOf []*jsonSchema
	anyOf []*jsonSchema
	allOf []*jsonSchema
	not   *jsonSchema
}

func (s *jsonSchema) AddEnum(i interface{}) error {

	is, err := marshalToString(i)
	if err != nil {
		return err
	}

	if isStringInSlice(s.enum, *is) {
		return errors.New("enum items must be unique")
	}

	s.enum = append(s.enum, *is)

	return nil
}

func (s *jsonSchema) AddOneOf(schema *jsonSchema) {
	s.oneOf = append(s.oneOf, schema)
}

func (s *jsonSchema) AddAllOf(schema *jsonSchema) {
	s.allOf = append(s.allOf, schema)
}

func (s *jsonSchema) AddAnyOf(schema *jsonSchema) {
	s.anyOf = append(s.anyOf, schema)
}

func (s *jsonSchema) SetNot(schema *jsonSchema) {
	s.not = schema
}

func (s *jsonSchema) HasEnum(i interface{}) (bool, error) {

	is, err := marshalToString(i)
	if err != nil {
		return false, err
	}

	return isStringInSlice(s.enum, *is), nil
}

func (s *jsonSchema) AddRequired(value string) error {

	if isStringInSlice(s.required, value) {
		return errors.New("required items must be unique")
	}

	s.required = append(s.required, value)

	return nil
}

func (s *jsonSchema) AddDefinitionChild(child *jsonSchema) {
	s.definitionsChildren = append(s.definitionsChildren, child)
}

func (s *jsonSchema) SetItemsChild(child *jsonSchema) {
	s.itemsChild = child
}

func (s *jsonSchema) AddPropertiesChild(child *jsonSchema) {
	s.propertiesChildren = append(s.propertiesChildren, child)
}

func (s *jsonSchema) HasProperty(name string) bool {

	for _, v := range s.propertiesChildren {
		if v.property == name {
			return true
		}
	}
	return false
}
