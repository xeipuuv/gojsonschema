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
// description      Utility functions for validation, type checking and cie.		
// 
// created          26-02-2013

package gojsonschema

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
