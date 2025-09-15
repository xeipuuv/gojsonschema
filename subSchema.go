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
// description      Defines the structure of a sub-subSchema.
//                  A sub-subSchema can contain other sub-schemas.
//
// created          27-02-2013

package gojsonschema

import (
	"encoding/json"
	"github.com/xeipuuv/gojsonreference"
	"math/big"
	"reflect"
	"regexp"
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

type subSchema struct {
	draft *Draft

	// basic subSchema meta properties
	id          *gojsonreference.JsonReference
	title       *string
	description *string

	property string

	// Quick pass/fail for boolean schemas
	pass *bool

	// Types associated with the subSchema
	types jsonSchemaType

	// Reference url
	ref *gojsonreference.JsonReference
	// Schema referenced
	refSchema *subSchema

	// hierarchy
	parent                      *subSchema
	itemsChildren               []*subSchema
	itemsChildrenIsSingleSchema bool
	propertiesChildren          []*subSchema

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
	patternProperties    map[string]*subSchema
	propertyNames        *subSchema

	// validation : array
	minItems    *int
	maxItems    *int
	uniqueItems bool
	contains    *subSchema

	additionalItems interface{}

	// validation : all
	_const *string //const is a golang keyword
	enum   []string

	// validation : subSchema
	oneOf []*subSchema
	anyOf []*subSchema
	allOf []*subSchema
	not   *subSchema
	_if   *subSchema // if/else are golang keywords
	_then *subSchema
	_else *subSchema
}

func (v *subSchema) MarshalJSON() ([]byte, error) {
	tmp := make(map[string]interface{})
	result := make(map[string]interface{})
	if len(v.types.types) > 0 {
		switch v.types.types[0] {
		//case "boolean", "null": // do nothing
		case "number", "integer":
			// type、multipleOf、minimum、exclusiveMinimum、maximum、exclusiveMaximum
			tmp = map[string]interface{}{
				"multipleOf":       ensureField(v.multipleOf),
				"minimum":          ensureField(v.minimum),
				"exclusiveMinimum": ensureField(v.exclusiveMinimum),
				"maximum":          ensureField(v.maximum),
				"exclusiveMaximum": ensureField(v.exclusiveMaximum),
			}
		case "string":
			// type、minLength、maxLength、pattern、format
			tmp = map[string]interface{}{
				"minLength": ensureField(v.minLength),
				"maxLength": ensureField(v.maxLength),
				"pattern":   ensureField(v.pattern),
				"format":    ensureField(v.format),
			}
		case "object":
			// type、properties、additionalProperties、required, propertyNames、minProperties、
			tmp = map[string]interface{}{
				"minProperties":        ensureField(v.minProperties),
				"maxProperties":        ensureField(v.maxProperties),
				"required":             ensureField(v.required),
				"dependencies":         ensureField(v.dependencies),
				"additionalProperties": ensureField(v.additionalProperties),
				"patternProperties":    ensureField(v.patternProperties),
				"propertyNames":        ensureField(v.propertyNames),
			}
			if v.propertiesChildren != nil {
				properties := make(map[string]interface{})
				for _, propertyChild := range v.propertiesChildren {
					properties[propertyChild.property] = propertyChild
				}
				tmp["properties"] = properties
			}
		case "array":
			tmp = map[string]interface{}{
				"minItems":        ensureField(v.minItems),
				"maxItems":        ensureField(v.maxItems),
				"uniqueItems":     ensureField(v.uniqueItems),
				"contains":        ensureField(v.contains),
				"additionalItems": ensureField(v.additionalItems),
			}
		}
		tmp["type"] = v.types.types[0]
		tmp["const"] = ensureField(v._const)
		tmp["enum"] = ensureField(v.enum)
	}
	for key, value := range tmp {
		if value != nil {
			result[key] = value
		}
	}
	return json.Marshal(result)
}

func ensureField(p interface{}) interface{} {
	if isNil(p) {
		return nil
	}
	var ret interface{}
	rv := reflect.ValueOf(p)
	if rv.Kind() == reflect.Ptr {
		if rv.Type() == reflect.TypeOf(&big.Rat{}) {
			ret = (p.(*big.Rat)).Num()
		} else if rv.Type() == reflect.TypeOf(&regexp.Regexp{}) {
			ret = (p.(*regexp.Regexp)).String()
		} else {
			switch rv.Elem().Kind() {
			case reflect.Int:
				ret = *(p.(*int))
			case reflect.String:
				ret = *(p.(*string))
			case reflect.Slice, reflect.Interface, reflect.Map:
				ret = p
			}
		}
	} else {
		switch rv.Kind() {
		case reflect.String:
			tmp := p.(string)
			if len(tmp) != 0 {
				ret = tmp
			}
		case reflect.Slice:
			ret = p
		}
	}
	return ret
}

func isNil(i interface{}) bool {
	// https://mangatmodi.medium.com/go-check-nil-interface-the-right-way-d142776edef1
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false

}
