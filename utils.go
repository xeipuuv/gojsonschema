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
// description      Various utility functions.
//
// created          26-02-2013

package gojsonschema

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
)

func isKind(what interface{}, kind reflect.Kind) bool {
	return reflect.ValueOf(what).Kind() == kind
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

// Practical when it comes to differentiate a float from an integer since JSON only knows numbers
// NOTE go's Parse(U)Int funcs accepts 1.0, 45.0 as integers
func isFloat64AnInteger(n float64) bool {
	_, errInt := strconv.ParseInt(fmt.Sprintf("%v", n), 10, 64)
	_, errUint := strconv.ParseUint(fmt.Sprintf("%v", n), 10, 64)
	return errInt == nil || errUint == nil
}

// formats a number so that it is displayed as the smallest string possible
func validationErrorFormatNumber(n float64) string {

	if isFloat64AnInteger(n) {

		valInt, errInt := strconv.ParseInt(fmt.Sprintf("%v", n), 10, 64)
		valUint, errUint := strconv.ParseUint(fmt.Sprintf("%v", n), 10, 64)

		if errInt == nil {
			return fmt.Sprintf("%v", valInt)
		} else if errUint == nil {
			return fmt.Sprintf("%v", valUint)
		}
	}

	return fmt.Sprintf("%g", n)
}

func marshalToJsonString(value interface{}) (*string, error) {

	mBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	sBytes := string(mBytes)
	return &sBytes, nil
}

const internalLogEnabled = false

func internalLog(message string) {
	if internalLogEnabled {
		log.Print(message)
	}
}
