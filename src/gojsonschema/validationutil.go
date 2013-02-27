// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"reflect"
)

const (
	ERROR_MESSAGE_MUST_BE_OF_TYPE = `%s must be of type %s`
	ERROR_MESSAGE_IS_REQUIRED     = `%s is required`
)

func isKind(what interface{}, kind reflect.Kind) bool {
	rValue := reflect.ValueOf(what)
	return rValue.Kind() == kind
}

func existsMapKey(m map[string]interface{}, k string) bool {
	_, ok := m[k]
	return ok
}

func isStringInSlice(s []string, what string) bool {
	for i := range s {
		if s[i] == what {
			return true
		}
	}
	return false
}
