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

func NewJsonSchema(document interface{}) (JsonSchema, error) {

	var s JsonSchema
	err := s.parse(document)
	return s, err

}

type JsonSchema struct {
	schemaKeyword      string
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
	if !mapValueExists(schemaMap, "$schema") {
		return errors.New(fmt.Sprintf(IS_REQUIRED, "$schema"))
	}

	if !interfaceOfKind(schemaMap["$schema"], reflect.String) {
		return errors.New(fmt.Sprintf(MUST_BE_STRING, "$schema"))
	}

	s.schemaKeyword = schemaMap["$schema"].(string)

	if !isStringInList(SUPPORTED_SCHEMA_KEYWORDS, s.schemaKeyword) {
		return errors.New(fmt.Sprintf(ENUM_VALUE_NOT_ALLOWED, "$schema", s.schemaKeyword))
	}

	// type
	if !mapValueExists(schemaMap, "type") {
		return errors.New(fmt.Sprintf(IS_REQUIRED, "type"))
	}

	if !interfaceOfKind(schemaMap["type"], reflect.String) {
		return errors.New(fmt.Sprintf(MUST_BE_STRING, "type"))
	}

	s.typeKeyword = schemaMap["type"].(string)

	if !isStringInList(SUPPORTED_TYPE_KEYWORDS, s.typeKeyword) {
		return errors.New(fmt.Sprintf(ENUM_VALUE_NOT_ALLOWED, "type", s.typeKeyword))
	}

	// title

	if mapValueExists(schemaMap, "title") {
		if !interfaceOfKind(schemaMap["title"], reflect.String) {
			return errors.New(fmt.Sprintf(MUST_BE_STRING, "title"))
		}
		s.titleKeyword = schemaMap["title"].(string)
	}

	// description

	if mapValueExists(schemaMap, "description") {
		if !interfaceOfKind(schemaMap["description"], reflect.String) {
			return errors.New(fmt.Sprintf(MUST_BE_STRING, "description"))
		}
		s.descriptionKeyword = schemaMap["description"].(string)
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
