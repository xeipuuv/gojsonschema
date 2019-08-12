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
// description      Various utility functions.
//
// created          26-02-2013

package gojsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
)

func isKind(what interface{}, kinds ...reflect.Kind) bool {
	target := what
	if isJSONNumber(what) {
		// JSON Numbers are strings!
		target = *mustBeNumber(what)
	}
	targetKind := reflect.ValueOf(target).Kind()
	for _, kind := range kinds {
		if kind == reflect.Map && isBsonD(target) {
			return true
		}
		if kind == reflect.Slice && isBsonD(target) {
			continue
		}
		if targetKind == kind {
			return true
		}
	}
	return false
}

func isBsonD(what interface{}) bool {
	_, ok := what.(bson.D)
	return ok
}

func mustBeMap(node interface{}) (map[string]interface{}, error) {
	switch n := node.(type) {
	case map[string]interface{}:
		return n, nil
	case bson.D:
		return n.Map(), nil
	default:
		return nil, errors.New(formatErrorDescription(
			Locale.ParseError(),
			ErrorDetails{
				"expected": []string{TYPE_OBJECT, "bson.D"},
			},
		))
	}

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

// indexStringInSlice returns the index of the first instance of 'what' in s or -1 if it is not found in s.
func indexStringInSlice(s []string, what string) int {
	for i := range s {
		if s[i] == what {
			return i
		}
	}
	return -1
}

func marshalToJSONString(value interface{}) (*string, error) {

	mBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	sBytes := string(mBytes)
	return &sBytes, nil
}

func marshalWithoutNumber(value interface{}) (*string, error) {

	// The JSON is decoded using https://golang.org/pkg/encoding/json/#Decoder.UseNumber
	// This means the numbers are internally still represented as strings and therefore 1.00 is unequal to 1
	// One way to eliminate these differences is to decode and encode the JSON one more time without Decoder.UseNumber
	// so that these differences in representation are removed

	jsonString, err := marshalToJSONString(value)
	if err != nil {
		return nil, err
	}

	var document interface{}

	err = json.Unmarshal([]byte(*jsonString), &document)
	if err != nil {
		return nil, err
	}

	return marshalToJSONString(document)
}

func isJSONNumber(what interface{}) bool {

	switch what.(type) {

	case json.Number:
		return true
	}

	return false
}

func isNumber(what interface{}) bool {
	switch what.(type) {
	case int, int8, int16, int32, int64, float32, float64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func isInteger(what interface{}) bool {
	switch what.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func checkJSONInteger(what interface{}) (isInt bool) {

	jsonNumber := what.(json.Number)

	bigFloat, isValidNumber := new(big.Rat).SetString(string(jsonNumber))

	return isValidNumber && bigFloat.IsInt()

}

// same as ECMA Number.MAX_SAFE_INTEGER and Number.MIN_SAFE_INTEGER
const (
	maxJSONFloat = float64(1<<53 - 1)  // 9007199254740991.0 	 2^53 - 1
	minJSONFloat = -float64(1<<53 - 1) //-9007199254740991.0	-2^53 - 1
)

func mustBeInteger(what interface{}) *int {

	if isJSONNumber(what) {

		number := what.(json.Number)

		isInt := checkJSONInteger(number)

		if isInt {

			int64Value, err := number.Int64()
			if err != nil {
				return nil
			}

			int32Value := int(int64Value)
			return &int32Value
		}

	} else if isInteger(what) {
		res, err := strconv.ParseInt(fmt.Sprintf("%d", what), 10, 64)
		if err != nil {
			return nil
		}
		resInt := int(res)
		return &resInt
	}

	return nil
}

func mustBeNumber(what interface{}) *big.Rat {

	if isJSONNumber(what) {
		number := what.(json.Number)
		float64Value, success := new(big.Rat).SetString(string(number))
		if success {
			return float64Value
		}

	} else if isNumber(what) {
		return mustBeGoNumber(what)
	}

	return nil
}

func mustBeGoNumber(what interface{}) *big.Rat {

	var float64Value *big.Rat
	var success bool
	switch n := what.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		float64Value, success = new(big.Rat).SetString(fmt.Sprintf("%d", n))
	case float32, float64:
		float64Value, success = new(big.Rat).SetString(fmt.Sprintf("%f", n))
	default:
		return nil
	}

	if success {
		return float64Value
	}
	return nil
}

func convertDocumentNode(val interface{}) interface{} {

	if lval, ok := val.([]interface{}); ok {

		res := []interface{}{}
		for _, v := range lval {
			res = append(res, convertDocumentNode(v))
		}

		return res

	}

	if mval, ok := val.(map[interface{}]interface{}); ok {

		res := map[string]interface{}{}

		for k, v := range mval {
			res[k.(string)] = convertDocumentNode(v)
		}

		return res

	}

	return val
}
