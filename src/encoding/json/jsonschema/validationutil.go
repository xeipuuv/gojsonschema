// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Utility functions for validation, type checking and cie.		
// 
// created      	26-02-2013

package jsonschema

import (
	"encoding/json"
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

func marshalToString(value interface{}) (*string, error) {
	mBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	sBytes := string(mBytes)
	return &sBytes, nil
}
