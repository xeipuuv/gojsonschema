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
// description      Contains consts, types and error messages.		
// 
// created          28-02-2013

package gojsonschema

import ()

const (
	KEY_SCHEMA            = "$schema"
	KEY_ID                = "$id"
	KEY_REF               = "$ref"
	KEY_TITLE             = "title"
	KEY_DESCRIPTION       = "description"
	KEY_TYPE              = "type"
	KEY_ITEMS             = "items"
	KEY_PROPERTIES        = "properties"
	KEY_MULTIPLE_OF       = "multipleOf"
	KEY_MINIMUM           = "minimum"
	KEY_MAXIMUM           = "maximum"
	KEY_EXCLUSIVE_MINIMUM = "exclusiveMinimum"
	KEY_EXCLUSIVE_MAXIMUM = "exclusiveMaximum"
	KEY_MIN_LENGTH        = "minLength"
	KEY_MAX_LENGTH        = "maxLength"
	KEY_PATTERN           = "pattern"
	KEY_MIN_PROPERTIES    = "minProperties"
	KEY_MAX_PROPERTIES    = "maxProperties"
	KEY_REQUIRED          = "required"
	KEY_MIN_ITEMS         = "minItems"
	KEY_MAX_ITEMS         = "maxItems"
	KEY_UNIQUE_ITEMS      = "uniqueItems"
	KEY_ENUM              = "enum"
	KEY_ONE_OF            = "oneOf"
	KEY_ANY_OF            = "anyOf"
	KEY_ALL_OF            = "allOf"
	KEY_NOT               = "not"

	STRING_STRING           = "string"
	STRING_ARRAY_OF_STRINGS = "array of strings"
	STRING_OBJECT           = "object"
	STRING_SCHEMA           = "schema"
	STRING_PROPERTIES       = "properties"

	ROOT_SCHEMA_PROPERTY = "(root)"
)

const (
	ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y = `%s must be of type %s`
	ERROR_MESSAGE_X_IS_REQUIRED       = `%s is required`
)

const (
	TYPE_ARRAY   = `array`
	TYPE_BOOLEAN = `boolean`
	TYPE_INTEGER = `integer`
	TYPE_NUMBER  = `number`
	TYPE_NULL    = `null`
	TYPE_OBJECT  = `object`
	TYPE_STRING  = `string`
)

var JSON_TYPES []string
var SCHEMA_TYPES []string

func init() {
	JSON_TYPES = []string{
		TYPE_ARRAY,
		TYPE_BOOLEAN,
		TYPE_INTEGER,
		TYPE_NUMBER,
		TYPE_NULL,
		TYPE_OBJECT,
		TYPE_STRING}

	SCHEMA_TYPES = []string{
		TYPE_ARRAY,
		TYPE_BOOLEAN,
		TYPE_INTEGER,
		TYPE_NUMBER,
		TYPE_OBJECT,
		TYPE_STRING}
}
