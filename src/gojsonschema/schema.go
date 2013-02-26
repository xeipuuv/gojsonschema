// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"reflect"
)

const (
	JSON_SCHEMA_MUST_BE_OBJECT = `Schema root must be an object`
	MUST_BE_STRING             = `'%s' must be a string`
	IS_REQUIRED                = `'%s' is required`
	ENUM_VALUE_NOT_ALLOWED     = `%s '%s' is not allowed`
)

const (
	KEY_SCHEMA      = `$schema`
	KEY_ID          = `id`
	KEY_TITLE       = `title`
	KEY_DESCRIPTION = `description`
	KEY_TYPE        = `type`
)

func NewJsonSchema(document interface{}) (JsonSchema, error) {

	var s JsonSchema
	err := s.parse(document)
	return s, err

}

type JsonSchema struct {
	schemaKeyword      string
	idKeyword          string
	titleKeyword       string
	descriptionKeyword string
	typeKeyword        string

	document interface{}
}

func (s *JsonSchema) parse(document interface{}) error {

	s.document = document

	if !interfaceOfKind(s.document, reflect.Map) {
		return errors.New(JSON_SCHEMA_MUST_BE_OBJECT)
	}

	schemaMap := s.document.(map[string]interface{})

	// $schema
	if !mapValueExists(schemaMap, KEY_SCHEMA) {
		return errors.New(fmt.Sprintf(IS_REQUIRED, KEY_SCHEMA))
	}

	if !interfaceOfKind(schemaMap[KEY_SCHEMA], reflect.String) {
		return errors.New(fmt.Sprintf(MUST_BE_STRING, KEY_SCHEMA))
	}

	s.schemaKeyword = schemaMap[KEY_SCHEMA].(string)

	if !isStringInList(SUPPORTED_SCHEMA_KEYWORDS, s.schemaKeyword) {
		return errors.New(fmt.Sprintf(ENUM_VALUE_NOT_ALLOWED, KEY_SCHEMA, s.schemaKeyword))
	}

	// type
	if !mapValueExists(schemaMap, KEY_TYPE) {
		return errors.New(fmt.Sprintf(IS_REQUIRED, KEY_TYPE))
	}

	if !interfaceOfKind(schemaMap[KEY_TYPE], reflect.String) {
		return errors.New(fmt.Sprintf(MUST_BE_STRING, KEY_TYPE))
	}

	s.typeKeyword = schemaMap[KEY_TYPE].(string)

	if !isStringInList(SUPPORTED_TYPE_KEYWORDS, s.typeKeyword) {
		return errors.New(fmt.Sprintf(ENUM_VALUE_NOT_ALLOWED, KEY_TYPE, s.typeKeyword))
	}

	// id

	if mapValueExists(schemaMap, KEY_ID) {
		if !interfaceOfKind(schemaMap[KEY_ID], reflect.String) {
			return errors.New(fmt.Sprintf(MUST_BE_STRING, KEY_ID))
		}
		s.idKeyword = schemaMap[KEY_ID].(string)
	}

	// title

	if mapValueExists(schemaMap, KEY_TITLE) {
		if !interfaceOfKind(schemaMap[KEY_TITLE], reflect.String) {
			return errors.New(fmt.Sprintf(MUST_BE_STRING, KEY_TITLE))
		}
		s.titleKeyword = schemaMap[KEY_TITLE].(string)
	}

	// description

	if mapValueExists(schemaMap, KEY_DESCRIPTION) {
		if !interfaceOfKind(schemaMap[KEY_DESCRIPTION], reflect.String) {
			return errors.New(fmt.Sprintf(MUST_BE_STRING, KEY_DESCRIPTION))
		}
		s.descriptionKeyword = schemaMap[KEY_DESCRIPTION].(string)
	}

	return nil
}

// Schema validation funcs

var SUPPORTED_SCHEMA_KEYWORDS []string
var SUPPORTED_TYPE_KEYWORDS []string

func init() {
	SUPPORTED_SCHEMA_KEYWORDS = []string{
		`http://json-schema.org/schema#`,
		//		`http://json-schema.org/hyper-schema#`,
		`http://json-schema.org/draft-04/schema#`,
		//		`http://json-schema.org/draft-04/hyper-schema#`,
		//		`http://json-schema.org/draft-03/schema#`,
		//		`http://json-schema.org/draft-03/hyper-schema#`,
	}
	SUPPORTED_TYPE_KEYWORDS = []string{
		`array`,
		`boolean`,
		`integer`,
		`number`,
		`null`,
		`object`,
		`string`,
	}
}

func interfaceOfKind(node interface{}, kind reflect.Kind) bool {
	return reflect.ValueOf(node).Kind() == kind
}

func mapValueExists(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

func isStringInList(list []string, keyword string) bool {
	for _, v := range list {
		if v == keyword {
			return true
		}
	}
	return false
}
