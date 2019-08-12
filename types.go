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
// description      Contains const types for schema and JSON.
//
// created          28-02-2013

package gojsonschema

// Type constants
const (
	TYPE_ARRAY   	 = `array`
	TYPE_BOOLEAN 	 = `boolean`
	TYPE_INTEGER 	 = `integer`
	TYPE_NUMBER  	 = `number`
	TYPE_NULL    	 = `null`
	TYPE_OBJECT  	 = `object`
	TYPE_STRING  	 = `string`
	TYPE_DOUBLE  	 = `double`
	TYPE_OBJECT_ID 	 = `objectId`
	TYPE_BOOL		 = `bool`
	TYPE_DATE 		 = `date`
	TYPE_REGEX 		 = `regex`
	TYPE_TIMESTAMP 	 = `timestamp`
	TYPE_INT32 		 = `int`
	TYPE_INT64 		 = `long`
	TYPE_DECIMAL128  = `decimal`
)

// JSON_TYPES hosts the list of type that are supported in JSON
var JSON_TYPES []string

// SCHEMA_TYPES hosts the list of type that are supported in schemas
var SCHEMA_TYPES []string
var BSON_TYPES []string

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

	BSON_TYPES = []string{
		TYPE_STRING,
		TYPE_OBJECT,
		TYPE_ARRAY,
		TYPE_OBJECT_ID,
		TYPE_BOOLEAN,
		TYPE_BOOL,
		TYPE_NULL,
		TYPE_REGEX,
		TYPE_DATE,
		TYPE_TIMESTAMP,
		TYPE_INT32,
		TYPE_INT64,
		TYPE_DECIMAL128,
		TYPE_DOUBLE,
		TYPE_NUMBER}
}
