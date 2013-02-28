// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"fmt"
	"reflect"
	"strconv"
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

func isFloat64AnInteger(n float64) bool {
	_, err := strconv.Atoi(fmt.Sprintf("%v", n))
	return err == nil
}
