// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"reflect"
)

const (
	ERROR_MESSAGE_MUST_BE_OF_TYPE = `%s must be of type %s`
)

func isKind(what interface{}, kind reflect.Kind) bool {
	rValue := reflect.ValueOf(what)
	return rValue.Kind() == kind
}

func existsMapKey(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}
