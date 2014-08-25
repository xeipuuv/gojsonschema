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
	"math"
	"reflect"
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

const (
	bias = 1023
)

// http://blog.labix.org/2013/06/27/ieee-754-brain-teaser
// http://www.goinggo.net/2013/08/gustavos-ieee-754-brain-teaser.html
func isFloat64AnInteger(f float64) bool {

	if f == 0 {
		return true
	}

	if math.IsNaN(f) {
		return false
	} else if math.IsInf(f, 0) {
		return false
	}

	bits := math.Float64bits(f)
	//sign := bits & (1 << 63)
	frac := (bits & ((1 << 52) - 1)) | (1 << 52)
	exp := int((bits>>52)&0x7ff) - bias - 52

	if exp < -52 || exp < 0 && (frac&(1<<uint64(-exp)-1)) != 0 {
		return false
	}

	return true
}

// formats a number so that it is displayed as the smallest string possible
func validationErrorFormatNumber(n float64) string {

	if isFloat64AnInteger(n) {
		return fmt.Sprintf("%d", int64(n))
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
