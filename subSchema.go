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
// description      Defines the structure of a sub-SubSchema.
//                  A sub-SubSchema can contain other sub-schemas.
//
// created          27-02-2013

package gojsonschema

import (
	"errors"
	"math/big"
	"regexp"
	"strings"

	"github.com/xeipuuv/gojsonreference"
)

// Constants
const (
	KEY_SCHEMA                = "$schema"
	KEY_ID                    = "id"
	KEY_ID_NEW                = "$id"
	KEY_REF                   = "$ref"
	KEY_TITLE                 = "title"
	KEY_DESCRIPTION           = "description"
	KEY_TYPE                  = "type"
	KEY_ITEMS                 = "items"
	KEY_ADDITIONAL_ITEMS      = "additionalItems"
	KEY_PROPERTIES            = "properties"
	KEY_PATTERN_PROPERTIES    = "patternProperties"
	KEY_ADDITIONAL_PROPERTIES = "additionalProperties"
	KEY_PROPERTY_NAMES        = "propertyNames"
	KEY_DEFINITIONS           = "definitions"
	KEY_MULTIPLE_OF           = "multipleOf"
	KEY_MINIMUM               = "minimum"
	KEY_MAXIMUM               = "maximum"
	KEY_EXCLUSIVE_MINIMUM     = "exclusiveMinimum"
	KEY_EXCLUSIVE_MAXIMUM     = "exclusiveMaximum"
	KEY_MIN_LENGTH            = "minLength"
	KEY_MAX_LENGTH            = "maxLength"
	KEY_PATTERN               = "pattern"
	KEY_FORMAT                = "format"
	KEY_MIN_PROPERTIES        = "minProperties"
	KEY_MAX_PROPERTIES        = "maxProperties"
	KEY_DEPENDENCIES          = "dependencies"
	KEY_REQUIRED              = "required"
	KEY_MIN_ITEMS             = "minItems"
	KEY_MAX_ITEMS             = "maxItems"
	KEY_UNIQUE_ITEMS          = "uniqueItems"
	KEY_CONTAINS              = "contains"
	KEY_CONST                 = "const"
	KEY_ENUM                  = "enum"
	KEY_ONE_OF                = "oneOf"
	KEY_ANY_OF                = "anyOf"
	KEY_ALL_OF                = "allOf"
	KEY_NOT                   = "not"
	KEY_IF                    = "if"
	KEY_THEN                  = "then"
	KEY_ELSE                  = "else"
)

type SubSchema struct {
	draft *Draft

	base *Schema

	node interface{}

	// basic SubSchema meta properties
	id          *gojsonreference.JsonReference
	title       *string
	description *string

	property string

	// Types associated with the SubSchema
	types jsonSchemaType

	// Reference url
	ref *gojsonreference.JsonReference
	// Schema referenced
	refSchema *SubSchema

	// hierarchy
	parent                      *SubSchema
	itemsChildren               []*SubSchema
	itemsChildrenIsSingleSchema bool
	propertiesChildren          []*SubSchema

	// validation : number / integer
	multipleOf       *big.Rat
	maximum          *big.Rat
	exclusiveMaximum *big.Rat
	minimum          *big.Rat
	exclusiveMinimum *big.Rat

	// validation : string
	minLength *int
	maxLength *int
	pattern   *regexp.Regexp
	format    string

	// validation : object
	minProperties *int
	maxProperties *int
	required      []string

	dependencies         map[string]interface{}
	additionalProperties interface{}
	patternProperties    map[string]*SubSchema
	propertyNames        *SubSchema

	// validation : array
	minItems    *int
	maxItems    *int
	uniqueItems bool
	contains    *SubSchema

	additionalItems interface{}

	// validation : all
	_const *string // const is a golang keyword
	enum   []string

	// validation : SubSchema
	oneOf []*SubSchema
	anyOf []*SubSchema
	allOf []*SubSchema
	not   *SubSchema
	_if   *SubSchema // if/else are golang keywords
	_then *SubSchema
	_else *SubSchema
}

func (s *SubSchema) Node() interface{} {
	return s.node
}

func (s *SubSchema) AddConst(i interface{}) error {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return err
	}
	s._const = is
	return nil
}

func (s *SubSchema) AddEnum(i interface{}) error {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return err
	}

	if isStringInSlice(s.enum, *is) {
		return errors.New(formatErrorDescription(
			Locale.KeyItemsMustBeUnique(),
			ErrorDetails{"key": KEY_ENUM},
		))
	}

	s.enum = append(s.enum, *is)

	return nil
}

func (s *SubSchema) ContainsEnum(i interface{}) (bool, error) {

	is, err := marshalWithoutNumber(i)
	if err != nil {
		return false, err
	}

	return isStringInSlice(s.enum, *is), nil
}

func (s *SubSchema) AddOneOf(subSchema *SubSchema) {
	s.oneOf = append(s.oneOf, subSchema)
}

func (s *SubSchema) AddAllOf(subSchema *SubSchema) {
	s.allOf = append(s.allOf, subSchema)
}

func (s *SubSchema) AddAnyOf(subSchema *SubSchema) {
	s.anyOf = append(s.anyOf, subSchema)
}

func (s *SubSchema) SetNot(subSchema *SubSchema) {
	s.not = subSchema
}

func (s *SubSchema) SetIf(subSchema *SubSchema) {
	s._if = subSchema
}

func (s *SubSchema) SetThen(subSchema *SubSchema) {
	s._then = subSchema
}

func (s *SubSchema) SetElse(subSchema *SubSchema) {
	s._else = subSchema
}

func (s *SubSchema) AddRequired(value string) error {

	if isStringInSlice(s.required, value) {
		return errors.New(formatErrorDescription(
			Locale.KeyItemsMustBeUnique(),
			ErrorDetails{"key": KEY_REQUIRED},
		))
	}

	s.required = append(s.required, value)

	return nil
}

func (s *SubSchema) AddItemsChild(child *SubSchema) {
	s.itemsChildren = append(s.itemsChildren, child)
}

func (s *SubSchema) AddPropertiesChild(child *SubSchema) {
	s.propertiesChildren = append(s.propertiesChildren, child)
}

func (s *SubSchema) PatternPropertiesString() string {

	if s.patternProperties == nil || len(s.patternProperties) == 0 {
		return STRING_UNDEFINED // should never happen
	}

	patternPropertiesKeySlice := []string{}
	for pk := range s.patternProperties {
		patternPropertiesKeySlice = append(patternPropertiesKeySlice, `"`+pk+`"`)
	}

	if len(patternPropertiesKeySlice) == 1 {
		return patternPropertiesKeySlice[0]
	}

	return "[" + strings.Join(patternPropertiesKeySlice, ",") + "]"

}
