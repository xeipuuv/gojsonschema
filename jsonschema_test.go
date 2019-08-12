// Copyright 2017 johandorland ( https://github.com/johandorland )
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

package gojsonschema

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type jsonSchemaTest struct {
	Description string `json:"description"`
	// Some tests may not always pass, so some tests are manually edited to include
	// an extra attribute whether that specific test should be disabled and skipped
	Disabled bool                 `json:"disabled"`
	Schema   interface{}          `json:"schema"`
	Tests    []jsonSchemaTestCase `json:"tests"`
}
type jsonSchemaTestCase struct {
	Description    string      `json:"description"`
	Data           interface{} `json:"data"`
	Valid          bool        `json:"valid"`
	PassValidation bool        `json:"passValidation"`
	ValidateTest   bool        `json:"validateTest"`
	Expression     interface{} `json:"expression"`
	FieldPath      []string    `json:"fieldPath"`
}

//Skip any directories not named appropiately
// filepath.Walk will also visit files in the root of the test directory
var testDirectories = regexp.MustCompile(`(draft\d+)`)
var draftMapping = map[string]Draft{
	"draft4": Draft4,
	"draft6": Draft6,
	"draft7": Draft7,
}

func executeTests(t *testing.T, path string) error {
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("Error (%s)\n", err.Error())
	}
	fmt.Println(file.Name())

	var tests []jsonSchemaTest
	d := json.NewDecoder(file)
	d.UseNumber()
	err = d.Decode(&tests)

	if err != nil {
		t.Errorf("Error (%s)\n", err.Error())
	}

	draft := Hybrid
	if m := testDirectories.FindString(path); m != "" {
		draft = draftMapping[m]
	}

	for _, test := range tests {
		fmt.Println("    " + test.Description)

		if test.Disabled {
			continue
		}

		testSchemaLoader := NewRawLoader(test.Schema)
		sl := NewSchemaLoader(NewNoopEvaluator())
		sl.Draft = draft
		sl.Validate = true
		testSchema, err := sl.Compile(testSchemaLoader)

		if err != nil {
			t.Errorf("Error (%s)\n", err.Error())
		}

		for _, testCase := range test.Tests {
			testDataLoader := NewRawLoader(testCase.Data)
			result, err := testSchema.Validate(testDataLoader)

			if err != nil {
				t.Errorf("Error (%s)\n", err.Error())
			}

			if result.Valid() != testCase.Valid {
				schemaString, _ := marshalToJSONString(test.Schema)
				testCaseString, _ := marshalToJSONString(testCase.Data)

				t.Errorf("Test failed : %s\n"+
					"%s.\n"+
					"%s.\n"+
					"expects: %t, given %t\n"+
					"Schema: %s\n"+
					"Data: %s\n",
					file.Name(),
					test.Description,
					testCase.Description,
					testCase.Valid,
					result.Valid(),
					*schemaString,
					*testCaseString)
			}
		}
	}
	return nil
}

func TestSuite(t *testing.T) {

	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	wd = filepath.Join(wd, "testdata")

	go func() {
		err := http.ListenAndServe(":1234", http.FileServer(http.Dir(filepath.Join(wd, "remotes"))))
		if err != nil {

			panic(err.Error())
		}
	}()

	err = filepath.Walk(wd, func(path string, fileInfo os.FileInfo, err error) error {
		if fileInfo.IsDir() && path != wd && !testDirectories.MatchString(fileInfo.Name()) {
			return filepath.SkipDir
		}
		if !strings.HasSuffix(fileInfo.Name(), ".json") {
			return nil
		}
		return executeTests(t, path)
	})
	if err != nil {
		t.Errorf("Error (%s)\n", err.Error())
	}
}

func TestFormats(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	wd = filepath.Join(wd, "testdata")

	dirs, err := ioutil.ReadDir(wd)

	if err != nil {
		panic(err.Error())
	}

	for _, dir := range dirs {
		if testDirectories.MatchString(dir.Name()) {
			formatsDirectory := filepath.Join(wd, dir.Name(), "optional", "format")
			err = filepath.Walk(formatsDirectory, func(path string, fileInfo os.FileInfo, err error) error {
				if fileInfo == nil || !strings.HasSuffix(fileInfo.Name(), ".json") {
					return nil
				}
				return executeTests(t, path)
			})

			if err != nil {
				t.Errorf("Error (%s)\n", err.Error())
			}
		}
	}
}

func TestBSONTypes(t *testing.T) {

	for _, test := range testCases() {
		testSchemaLoader := NewRawLoader(test.Schema)

		for _, testCase := range test.Tests {
			testDataLoader := NewGoLoader(testCase.Data)

			var testSchema *Schema
			var err error
			if testCase.ValidateTest {
				testSchema, err = NewSchema(testSchemaLoader, &MockValidateEvaluator{
					t:                  t,
					expectedExpression: testCase.Expression,
					expectedFieldPath:  testCase.FieldPath,
					valid:              testCase.PassValidation,
				})
			} else {
				testSchema, err = NewSchema(testSchemaLoader, NewNoopEvaluator())
			}
			if err != nil {
				t.Fatalf("Error (%s)\n(%s)\n", test.Description, err.Error())
			}

			result, err := testSchema.Validate(testDataLoader)

			if err != nil {
				t.Fatalf("Error (%s)\n", err.Error())
			}

			if result.Valid() != testCase.Valid {
				schemaString, _ := marshalToJSONString(test.Schema)
				testCaseString, _ := marshalToJSONString(testCase.Data)

				t.Fatalf("Test failed : %s\n"+
					"%s.\n"+
					"expects: %t, given %t\n"+
					"Schema: %s\n"+
					"Data: %s\n",
					test.Description,
					testCase.Description,
					testCase.Valid,
					result.Valid(),
					*schemaString,
					*testCaseString)
			}
		}
	}
}

type MockValidateEvaluator struct {
	t                  *testing.T
	expectedExpression interface{}
	expectedFieldPath  []string
	valid              bool
}

func (evaluator *MockValidateEvaluator) Evaluate(expression interface{}, fieldPath []string) error {
	if !reflect.DeepEqual(expression, evaluator.expectedExpression) {
		evaluator.t.Errorf("Test failed : \nexpected: %v\n actual: %v\n", evaluator.expectedExpression, expression)
	}
	if !reflect.DeepEqual(fieldPath, evaluator.expectedFieldPath) {
		evaluator.t.Errorf("Test failed : \nexpected: %v\n actual: %v\n", evaluator.expectedFieldPath, fieldPath)
	}
	if evaluator.valid {
		return nil
	}
	return fmt.Errorf("validation error")
}

func bsonTypeTestCase(inputType, matchType string, shouldMatch bool) jsonSchemaTestCase {
	data := getTestData(inputType)
	tc := jsonSchemaTestCase{
		Data:        data,
		Description: fmt.Sprintf("a %s is a %s", inputType, matchType),
		Valid:       shouldMatch,
	}
	if !shouldMatch {
		tc.Description = fmt.Sprintf("a %s is not a %s", inputType, matchType)
	}
	return tc
}

func bsonTestCase(description string, data interface{}, shouldMatch bool) jsonSchemaTestCase {
	return jsonSchemaTestCase{
		Data:        data,
		Description: description,
		Valid:       shouldMatch,
	}
}

func validateTestCase(description string, data interface{}, shouldMatch bool, validate bool, expectedExpression interface{}, expectedFieldPath []string) jsonSchemaTestCase {
	return jsonSchemaTestCase{
		Data:           data,
		Description:    description,
		Valid:          shouldMatch,
		Expression:     expectedExpression,
		FieldPath:      expectedFieldPath,
		ValidateTest:   true,
		PassValidation: validate,
	}
}

func getTestData(inputType string) interface{} {
	switch inputType {
	case TYPE_OBJECT_ID:
		return primitive.NewObjectID()
	case TYPE_INT32, TYPE_INT64:
		return 1
	case TYPE_DOUBLE:
		return 1.1
	case TYPE_STRING:
		return "foo"
	case TYPE_OBJECT:
		return map[string]interface{}{}
	case TYPE_ARRAY:
		return []interface{}{1, 2, 3}
	case TYPE_BOOL, TYPE_BOOLEAN:
		return true
	case TYPE_NULL:
		return nil
	case TYPE_REGEX:
		return primitive.Regex{}
	case TYPE_DATE:
		return time.Now()
	case TYPE_DECIMAL128:
		decimal, err := primitive.ParseDecimal128("1.5")
		if err != nil {
			panic(err)
		}
		return decimal
	case "bson.D":
		return bson.D{}
	case TYPE_TIMESTAMP:
		return primitive.Timestamp{123, 0}
	default:
		panic(fmt.Sprintf("%s is not a supported test type", inputType))
	}
}

func testCases() []jsonSchemaTest {
	validateExpression := bson.D{
		{"%function", bson.D{
			{"name", "func0"},
			{"arguments", []string{"%%value"}},
		}},
	}
	allOfMap := map[string]interface{}{"foo": primitive.Regex{}, "bar": int32(2)}
	allOfBson := bson.D{{"foo", primitive.Regex{}}, {"bar", 2}}

	return []jsonSchemaTest{
		{
			Description: "objectId type matches objectId",
			Schema:      map[string]interface{}{"bsonType": "objectId"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_OBJECT_ID, true),
				bsonTypeTestCase(TYPE_INT32, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_OBJECT_ID, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_OBJECT_ID, false),
			},
		},
		{
			Description: "double type matches double",
			Schema:      map[string]interface{}{"bsonType": "double"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_DOUBLE, true),
				bsonTypeTestCase(TYPE_STRING, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_DOUBLE, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_DOUBLE, false),
			},
		},
		{
			Description: "string type matches string",
			Schema:      map[string]interface{}{"bsonType": "string"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_STRING, true),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_STRING, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_STRING, false),
			},
		},
		{
			Description: "array type matches array",
			Schema:      map[string]interface{}{"bsonType": "array"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_ARRAY, true),
				bsonTypeTestCase(TYPE_BOOL, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_ARRAY, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_ARRAY, false),
				bsonTypeTestCase("bson.D", TYPE_ARRAY, false),
			},
		},
		{
			Description: "object type matches object",
			Schema:      map[string]interface{}{"bsonType": "object"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_OBJECT, true),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_OBJECT, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_OBJECT, false),
				bsonTypeTestCase("bson.D", TYPE_OBJECT, true),
			},
		},
		{
			Description: "bool type matches bool",
			Schema:      map[string]interface{}{"bsonType": "bool"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_BOOL, true),
				bsonTypeTestCase(TYPE_BOOLEAN, TYPE_BOOL, true),
				bsonTypeTestCase(TYPE_NULL, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_BOOL, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_BOOL, false),
				bsonTypeTestCase("bson.D", TYPE_BOOL, false),
			},
		},
		{
			Description: "date type matches date",
			Schema:      map[string]interface{}{"bsonType": "date"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_DATE, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_DATE, true),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_DATE, false),
				bsonTypeTestCase("bson.D", TYPE_DATE, false),
			},
		},
		{
			Description: "null type matches null",
			Schema:      map[string]interface{}{"bsonType": "null"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_NULL, true),
				bsonTypeTestCase(TYPE_REGEX, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_NULL, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_NULL, false),
				bsonTypeTestCase("bson.D", TYPE_NULL, false),
			},
		},
		{
			Description: "regex type matches regex",
			Schema:      map[string]interface{}{"bsonType": "regex"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_REGEX, true),
				bsonTypeTestCase(TYPE_DATE, TYPE_REGEX, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_REGEX, false),
				bsonTypeTestCase("bson.D", TYPE_REGEX, false),
			},
		},
		{
			Description: "int type matches int",
			Schema:      map[string]interface{}{"bsonType": "int"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_INT32, true),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_INT32, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_INT32, false),
				bsonTypeTestCase("bson.D", TYPE_INT32, false),
				bsonTypeTestCase(TYPE_TIMESTAMP, TYPE_INT32, false),
			},
		},
		{
			Description: "timestamp type matches timestamp",
			Schema:      map[string]interface{}{"bsonType": "timestamp"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_TIMESTAMP, false),
				bsonTypeTestCase("bson.D", TYPE_TIMESTAMP, false),
				bsonTypeTestCase(TYPE_TIMESTAMP, TYPE_TIMESTAMP, true),
			},
		},
		{
			Description: "long type matches long",
			Schema:      map[string]interface{}{"bsonType": "long"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_INT64, true),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_INT64, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_INT64, false),
				bsonTypeTestCase("bson.D", TYPE_INT64, false),
				bsonTypeTestCase(TYPE_TIMESTAMP, TYPE_INT64, false),
			},
		},
		{
			Description: "decimal type matches decimal",
			Schema:      map[string]interface{}{"bsonType": "decimal"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_STRING, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_DECIMAL128, true),
				bsonTypeTestCase("bson.D", TYPE_DECIMAL128, false),
				bsonTypeTestCase(TYPE_TIMESTAMP, TYPE_DECIMAL128, false),
			},
		},
		{
			Description: "number type matches number",
			Schema:      map[string]interface{}{"bsonType": "number"},
			Tests: []jsonSchemaTestCase{
				bsonTypeTestCase(TYPE_OBJECT_ID, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_INT32, TYPE_NUMBER, true),
				bsonTypeTestCase(TYPE_DOUBLE, TYPE_NUMBER, true),
				bsonTypeTestCase(TYPE_STRING, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_OBJECT, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_ARRAY, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_BOOL, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_NULL, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_REGEX, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_DATE, TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_DECIMAL128, TYPE_NUMBER, true),
				bsonTypeTestCase("bson.D", TYPE_NUMBER, false),
				bsonTypeTestCase(TYPE_TIMESTAMP, TYPE_NUMBER, false),
			},
		},
		{
			Description: "allOf with bson types",
			Schema: map[string]interface{}{"allOf": []interface{}{
				map[string]interface{}{
					"properties": map[string]interface{}{
						"bar": map[string]interface{}{
							"bsonType": TYPE_INT32,
						},
					},
					"required": []interface{}{"bar"},
				},
				map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": map[string]interface{}{
							"bsonType": TYPE_REGEX,
						},
					},
					"required": []interface{}{"foo"},
				},
			}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("matching types", map[string]interface{}{"foo": primitive.Regex{}, "bar": 2}, true),
				bsonTestCase("wrong type", map[string]interface{}{"foo": "baz", "bar": 2}, false),
				bsonTestCase("matching types with bson.D", bson.D{{"foo", primitive.Regex{}}, {"bar", 2}}, true),
				bsonTestCase("wrong type with bson.D", bson.D{{"foo", "baz"}, {"bar", 2}}, false),
			},
		},
		{
			Description: "anyOf with bson types",
			Schema: map[string]interface{}{"anyOf": []interface{}{
				map[string]interface{}{
					"bsonType": TYPE_OBJECT_ID,
				},
				map[string]interface{}{
					"bsonType": TYPE_ARRAY,
				},
			}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("matching bson type", primitive.NewObjectID(), true),
				bsonTestCase("matching array type", []interface{}{1, 2, 3}, true),
				bsonTestCase("no matching type", "foo", false),
			},
		},
		{
			Description: "oneOf with bson types",
			Schema: map[string]interface{}{"oneOf": []interface{}{
				map[string]interface{}{
					"bsonType": TYPE_INT32,
				},
				map[string]interface{}{
					"minimum": 2,
				},
			}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("matching bson type", 1, true),
				bsonTestCase("above minimum", 2.5, true),
				bsonTestCase("matching both", 3, false),
			},
		},
		{
			Description: "additionalItems as schema",
			Schema: map[string]interface{}{
				"items":           []interface{}{map[string]interface{}{}},
				"additionalItems": map[string]interface{}{"bsonType": TYPE_BOOL},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("additional items match schema", []interface{}{nil, true, false}, true),
				bsonTestCase("additional items do not match schema", []interface{}{nil, true, "hello"}, false),
			},
		},
		{
			Description: "additionalItems as schema with bson schema",
			Schema:      bson.D{{"items", []interface{}{bson.D{}}}, {"additionalItems", bson.D{{"bsonType", TYPE_BOOL}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("additional items match schema", []interface{}{nil, true, false}, true),
				bsonTestCase("additional items do not match schema", []interface{}{nil, true, "hello"}, false),
			},
		},
		{
			Description: "a schema given for items",
			Schema: map[string]interface{}{
				"items": map[string]interface{}{"bsonType": TYPE_DOUBLE},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("valid items", []interface{}{1.1, 2.1, 3.1}, true),
				bsonTestCase("wrong type of items", []interface{}{1.1, "x"}, false),
			},
		},
		{
			Description: "a schema given for items with bson schema",
			Schema:      bson.D{{"items", bson.D{{"bsonType", TYPE_DOUBLE}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("valid items", []interface{}{1.1, 2.1, 3.1}, true),
				bsonTestCase("wrong type of items", []interface{}{1.1, "x"}, false),
			},
		},
		{
			Description: "patternProperties validates properties matching a regex",
			Schema: map[string]interface{}{
				"patternProperties": map[string]interface{}{
					"f.*o": map[string]interface{}{"bsonType": TYPE_INT32},
				},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("a single valid match is valid", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("multiple valid matches is valid", map[string]interface{}{"foo": 1, "foooooo": 2}, true),
				bsonTestCase("a single invalid match is invalid", map[string]interface{}{"foo": "bar", "fooooo": 2}, false),
				bsonTestCase("a single valid match is valid with bson.D", bson.D{{"foo", 1}}, true),
				bsonTestCase("multiple valid matches is valid with bson.D", bson.D{{"foo", 1}, {"foooooo", 2}}, true),
				bsonTestCase("a single invalid match is invalid with bson.D", bson.D{{"foo", "bar"}, {"fooooo", 2}}, false),
			},
		},
		{
			Description: "patternProperties validates properties matching a regex with bson schema",
			Schema:      bson.D{{"patternProperties", bson.D{{"f.*o", bson.D{{"bsonType", TYPE_INT32}}}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("a single valid match is valid", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("multiple valid matches is valid", map[string]interface{}{"foo": 1, "foooooo": 2}, true),
				bsonTestCase("a single invalid match is invalid", map[string]interface{}{"foo": "bar", "fooooo": 2}, false),
				bsonTestCase("a single valid match is valid with bson.D", bson.D{{"foo", 1}}, true),
				bsonTestCase("multiple valid matches is valid with bson.D", bson.D{{"foo", 1}, {"foooooo", 2}}, true),
				bsonTestCase("a single invalid match is invalid with bson.D", bson.D{{"foo", "bar"}, {"fooooo", 2}}, false),
			},
		},
		{
			Description: "object properties validation",
			Schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"bsonType": TYPE_INT32},
					"bar": map[string]interface{}{"bsonType": TYPE_STRING},
				},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("both properties present and valid is valid", map[string]interface{}{"foo": 1, "bar": "baz"}, true),
				bsonTestCase("one property invalid is invalid", map[string]interface{}{"foo": 1, "bar": bson.D{}}, false),
				bsonTestCase("both properties present and valid is valid with bson.D", bson.D{{"foo", 1}, {"bar", "baz"}}, true),
				bsonTestCase("one property invalid is invalid with bson.D", bson.D{{"foo", 1}, {"bar", bson.D{}}}, false),
			},
		},
		{
			Description: "object properties validation with bson schema",
			Schema: bson.D{{
				"properties", bson.D{
					{"foo", bson.D{{"bsonType", TYPE_INT32}}},
					{"bar", bson.D{{"bsonType", TYPE_STRING}}},
				},
			}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("both properties present and valid is valid", map[string]interface{}{"foo": 1, "bar": "baz"}, true),
				bsonTestCase("one property invalid is invalid", map[string]interface{}{"foo": 1, "bar": bson.D{}}, false),
				bsonTestCase("both properties present and valid is valid with bson.D", bson.D{{"foo", 1}, {"bar", "baz"}}, true),
				bsonTestCase("one property invalid is invalid with bson.D", bson.D{{"foo", 1}, {"bar", bson.D{}}}, false),
			},
		},
		{
			Description: "with validate on base level",
			Schema: map[string]interface{}{
				"bsonType": "string",
				"validate": validateExpression,
			},
			Tests: []jsonSchemaTestCase{
				validateTestCase("passes when validate is true", "haley", true, true, validateExpression, []string{}),
				validateTestCase("does not pass when validate is false", "haley", false, false, validateExpression, []string{}),
			},
		},
		{
			Description: "with validate and multiple levels",
			Schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"bsonType": TYPE_STRING,
					},
					"info": map[string]interface{}{
						"bsonType": TYPE_OBJECT,
						"properties": map[string]interface{}{
							"id": map[string]interface{}{
								"bsonType": TYPE_OBJECT_ID,
							},
							"school": map[string]interface{}{
								"bsonType": TYPE_STRING,
								"validate": validateExpression,
							},
						},
					},
				},
			},
			Tests: []jsonSchemaTestCase{
				validateTestCase(
					"passes when validate is true",
					map[string]interface{}{
						"name": "haley",
						"info": map[string]interface{}{"id": primitive.NewObjectID(), "school": "UT Austin"},
					},
					true,
					true,
					validateExpression,
					[]string{"info", "school"},
				),
				validateTestCase(
					"does not pass when validate is false",
					map[string]interface{}{
						"name": "haley",
						"info": map[string]interface{}{"id": primitive.NewObjectID(), "school": "UT Austin"},
					},
					false,
					false,
					validateExpression,
					[]string{"info", "school"},
				),
				validateTestCase(
					"passes when validate is true with bson.D",
					bson.D{{"name", "haley"}, {"info", bson.D{{"id", primitive.NewObjectID()}, {"school", "UT Austin"}}}},
					true,
					true,
					validateExpression,
					[]string{"info", "school"},
				),
				validateTestCase(
					"does not pass when validate is false with bson.D",
					bson.D{{"name", "haley"}, {"info", bson.D{{"id", primitive.NewObjectID()}, {"school", "UT Austin"}}}},
					false,
					false,
					validateExpression,
					[]string{"info", "school"},
				),
			},
		},
		{
			Description: "with validate and allOf",
			Schema: map[string]interface{}{"allOf": []interface{}{
				map[string]interface{}{
					"properties": map[string]interface{}{
						"bar": map[string]interface{}{
							"bsonType": TYPE_INT32,
						},
					},
				},
				map[string]interface{}{
					"properties": map[string]interface{}{
						"foo": map[string]interface{}{
							"bsonType": TYPE_REGEX,
							"validate": validateExpression,
						},
					},
				},
			}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("passes when both are true", allOfMap, true, true, validateExpression, []string{"foo"}),
				validateTestCase("does not pass when validate is false", allOfMap, false, false, validateExpression, []string{"foo"}),
				validateTestCase(
					"does not pass when all are not true",
					map[string]interface{}{"foo": primitive.Regex{}, "bar": "hello"},
					false,
					true,
					validateExpression,
					[]string{"foo"},
				),
				validateTestCase("passes when both are true with bson.D", allOfBson, true, true, validateExpression, []string{"foo"}),
				validateTestCase("does not pass when validate is false with bson.D", allOfBson, false, false, validateExpression, []string{"foo"}),
				validateTestCase(
					"does not pass when all are not true with bson.D",
					bson.D{{"foo", primitive.Regex{}}, {"bar", "hello"}},
					false,
					true,
					validateExpression,
					[]string{"foo"},
				),
			},
		},
		{
			Description: "with validate and allOf and bson schema",
			Schema: bson.D{{"allOf", []interface{}{
				bson.D{{"properties", bson.D{{"bar", bson.D{{"bsonType", TYPE_INT32}}}}}},
				bson.D{{"properties", bson.D{{"foo", bson.D{{"bsonType", TYPE_REGEX}, {"validate", validateExpression}}}}}},
			}}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("passes when both are true", allOfMap, true, true, validateExpression, []string{"foo"}),
				validateTestCase("does not pass when validate is false", allOfMap, false, false, validateExpression, []string{"foo"}),
				validateTestCase(
					"does not pass when all are not true",
					map[string]interface{}{"foo": primitive.Regex{}, "bar": "hello"},
					false,
					true,
					validateExpression,
					[]string{"foo"},
				),
				validateTestCase("passes when both are true with bson.D", allOfBson, true, true, validateExpression, []string{"foo"}),
				validateTestCase("does not pass when validate is false with bson.D", allOfBson, false, false, validateExpression, []string{"foo"}),
				validateTestCase("does not pass when all are not true with bson.D",
					bson.D{{"foo", primitive.Regex{}}, {"bar", "hello"}},
					false,
					true,
					validateExpression,
					[]string{"foo"},
				),
			},
		},
		{
			Description: "with validate and anyOf",
			Schema: map[string]interface{}{"anyOf": []interface{}{
				map[string]interface{}{
					"bsonType": TYPE_OBJECT_ID,
				},
				map[string]interface{}{
					"bsonType": TYPE_ARRAY,
					"validate": validateExpression,
				},
			}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("passes when one is true", primitive.NewObjectID(), true, true, validateExpression, []string{}),
				validateTestCase("passes when one is true but validate on another is false", primitive.NewObjectID(), true, false, validateExpression, []string{}),
				validateTestCase("does not pass when validate is false", []interface{}{}, false, false, validateExpression, []string{}),
			},
		},
		{
			Description: "with validate and anyOf and bson schema",
			Schema: bson.D{{"anyOf", []interface{}{
				bson.D{{"bsonType", TYPE_OBJECT_ID}},
				bson.D{{"bsonType", TYPE_ARRAY}, {"validate", validateExpression}},
			}}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("passes when one is true", primitive.NewObjectID(), true, true, validateExpression, []string{}),
				validateTestCase("passes when one is true but validate on another is false", primitive.NewObjectID(), true, false, validateExpression, []string{}),
				validateTestCase("does not pass when validate is false", []interface{}{}, false, false, validateExpression, []string{}),
			},
		},
		{
			Description: "oneOf with bson types",
			Schema: map[string]interface{}{"oneOf": []interface{}{
				map[string]interface{}{
					"bsonType": TYPE_INT32,
				},
				map[string]interface{}{
					"minimum":  2,
					"validate": validateExpression,
				},
			}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("matching bson type", 1, true, true, validateExpression, []string{}),
				validateTestCase("above minimum", 2.5, true, true, validateExpression, []string{}),
				validateTestCase("above minimum but fail validate", 2.5, false, false, validateExpression, []string{}),
				validateTestCase("matching both", 3, false, true, validateExpression, []string{}),
			},
		},
		{
			Description: "oneOf with bson types and bson schema",
			Schema: bson.D{{"oneOf", []interface{}{
				bson.D{{"bsonType", TYPE_INT32}},
				bson.D{{"minimum", 2}, {"validate", validateExpression}},
			}}},
			Tests: []jsonSchemaTestCase{
				validateTestCase("matching bson type", 1, true, true, validateExpression, []string{}),
				validateTestCase("above minimum", 2.5, true, true, validateExpression, []string{}),
				validateTestCase("above minimum but fail validate", 2.5, false, false, validateExpression, []string{}),
				validateTestCase("matching both", 3, false, true, validateExpression, []string{}),
			},
		},
		{
			Description: "additionalProperties as a map",
			Schema: map[string]interface{}{
				"properties":           map[string]interface{}{"foo": map[string]interface{}{}, "bar": map[string]interface{}{}},
				"additionalProperties": map[string]interface{}{"bsonType": TYPE_BOOL},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("no additional properties is valid", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("an additional valid property is valid", map[string]interface{}{"foo": 1, "bar": 2, "quux": true}, true),
				bsonTestCase("an additional invalid property is invalid", map[string]interface{}{"foo": 1, "bar": 2, "quux": 12}, false),
				bsonTestCase("no additional properties is valid as bson", bson.D{{"foo", 1}}, true),
				bsonTestCase("an additional valid property is valid as bson", bson.D{{"foo", 1}, {"bar", 2}, {"quux", true}}, true),
				bsonTestCase("an additional invalid property is invalid as bson", bson.D{{"foo", 1}, {"bar", 2}, {"quux", 12}}, false),
			},
		},
		{
			Description: "additionalProperties as a bson.D",
			Schema: bson.D{
				{"properties", bson.D{{"foo", bson.D{}}, {"bar", bson.D{}}}},
				{"additionalProperties", bson.D{{"bsonType", TYPE_BOOL}}},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("no additional properties is valid", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("an additional valid property is valid", map[string]interface{}{"foo": 1, "bar": 2, "quux": true}, true),
				bsonTestCase("an additional invalid property is invalid", map[string]interface{}{"foo": 1, "bar": 2, "quux": 12}, false),
				bsonTestCase("no additional properties is valid as bson", bson.D{{"foo", 1}}, true),
				bsonTestCase("an additional valid property is valid as bson", bson.D{{"foo", 1}, {"bar", 2}, {"quux", true}}, true),
				bsonTestCase("an additional invalid property is invalid as bson", bson.D{{"foo", 1}, {"bar", 2}, {"quux", 12}}, false),
			},
		},
		{
			Description: "boolean schema works with true",
			Schema:      true,
			Tests: []jsonSchemaTestCase{
				bsonTestCase("number is valid", 1, true),
			},
		},
		{
			Description: "boolean schema works with false",
			Schema:      false,
			Tests: []jsonSchemaTestCase{
				bsonTestCase("number is valid", 1, false),
			},
		},
		{
			Description: "contains keyword validation with map schema",
			Schema:      map[string]interface{}{"contains": map[string]interface{}{"minimum": 5}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("array with item matching schema (5) is valid", []interface{}{3, 4, 5}, true),
				bsonTestCase("array without items matching schema is invalid", []interface{}{2, 3, 4}, false),
				bsonTestCase("empty array is invalid", []interface{}{}, false),
				bsonTestCase("not array is valid", map[string]interface{}{}, true),
				bsonTestCase("not array is valid", bson.D{}, true),
			},
		},
		{
			Description: "contains keyword validation with bson schema",
			Schema:      bson.D{{"contains", bson.D{{"minimum", 5}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("array with item matching schema (5) is valid", []interface{}{3, 4, 5}, true),
				bsonTestCase("array without items matching schema is invalid", []interface{}{2, 3, 4}, false),
				bsonTestCase("empty array is invalid", []interface{}{}, false),
				bsonTestCase("not array is valid", map[string]interface{}{}, true),
				bsonTestCase("not array is valid", bson.D{}, true),
			},
		},
		{
			Description: "dependencies with map schema",
			Schema:      map[string]interface{}{"dependencies": map[string]interface{}{"bar": []interface{}{"foo"}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("nondependant", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("with dependency", map[string]interface{}{"foo": 1, "bar": 2}, true),
				bsonTestCase("missing dependency", map[string]interface{}{"bar": 2}, false),
				bsonTestCase("nondependant with bson", bson.D{{"foo", 1}}, true),
				bsonTestCase("with dependency with bson", bson.D{{"foo", 1}, {"bar", 2}}, true),
				bsonTestCase("missing dependency with bson", bson.D{{"bar", 2}}, false),
				bsonTestCase("ignores arrays", []interface{}{"bar"}, true),
			},
		},
		{
			Description: "dependencies with bson schema",
			Schema:      bson.D{{"dependencies", bson.D{{"bar", []interface{}{"foo"}}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("nondependant", map[string]interface{}{"foo": 1}, true),
				bsonTestCase("with dependency", map[string]interface{}{"foo": 1, "bar": 2}, true),
				bsonTestCase("missing dependency", map[string]interface{}{"bar": 2}, false),
				bsonTestCase("nondependant with bson", bson.D{{"foo", 1}}, true),
				bsonTestCase("with dependency with bson", bson.D{{"foo", 1}, {"bar", 2}}, true),
				bsonTestCase("missing dependency with bson", bson.D{{"bar", 2}}, false),
				bsonTestCase("ignores arrays", []interface{}{"bar"}, true),
			},
		},
		{
			Description: "simple enum validation with map schema",
			Schema:      map[string]interface{}{"enum": []interface{}{1, 2, 3}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("one of the enum is valid", 1, true),
				bsonTestCase("something else is invalid", 4, false),
			},
		},
		{
			Description: "simple enum validation with bson schema",
			Schema:      bson.D{{"enum", []interface{}{1, 2, 3}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("one of the enum is valid", 1, true),
				bsonTestCase("something else is invalid", 4, false),
			},
		},
		{
			Description: "validate against correct branch, then vs else with map schema",
			Schema: map[string]interface{}{
				"if":   map[string]interface{}{"exclusiveMaximum": 0},
				"then": map[string]interface{}{"minimum": -10},
				"else": map[string]interface{}{"multipleOf": 2},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("valid through then", -1, true),
				bsonTestCase("invalid through then", -100, false),
				bsonTestCase("valid through else", 4, true),
				bsonTestCase("invalid through else", 3, false),
			},
		},
		{
			Description: "validate against correct branch, then vs else with bson schema",
			Schema: bson.D{
				{"if", bson.D{{"exclusiveMaximum", 0}}},
				{"then", bson.D{{"minimum", -10}}},
				{"else", bson.D{{"multipleOf", 2}}},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("valid through then", -1, true),
				bsonTestCase("invalid through then", -100, false),
				bsonTestCase("valid through else", 4, true),
				bsonTestCase("invalid through else", 3, false),
			},
		},
		{
			Description: "not with map schema",
			Schema:      map[string]interface{}{"not": map[string]interface{}{"bsonType": TYPE_INT32}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("allowed", "foo", true),
				bsonTestCase("disallowed", 1, false),
			},
		},
		{
			Description: "not with bson schema",
			Schema:      bson.D{{"not", bson.D{{"bsonType", TYPE_INT32}}}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("allowed", "foo", true),
				bsonTestCase("disallowed", 1, false),
			},
		},
		{
			Description: "relative pointer ref to object with map schema",
			Schema: map[string]interface{}{
				"properties": map[string]interface{}{
					"foo": map[string]interface{}{"bsonType": TYPE_INT32},
					"bar": map[string]interface{}{"$ref": "#/properties/foo"},
				},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("match", map[string]interface{}{"bar": 3}, true),
				bsonTestCase("mismatch", map[string]interface{}{"bar": true}, false),
				bsonTestCase("match with bson", bson.D{{"bar", 3}}, true),
				bsonTestCase("mismatch with bson", bson.D{{"bar", true}}, false),
			},
		},
		{
			Description: "relative pointer ref to object with bson schema",
			Schema: bson.D{
				{"properties", bson.D{
					{"foo", bson.D{{"bsonType", TYPE_INT32}}},
					{"bar", bson.D{{"$ref", "#/properties/foo"}}},
				}},
			},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("match", map[string]interface{}{"bar": 3}, true),
				bsonTestCase("mismatch", map[string]interface{}{"bar": true}, false),
				bsonTestCase("match with bson", bson.D{{"bar", 3}}, true),
				bsonTestCase("mismatch with bson", bson.D{{"bar", true}}, false),
			},
		},
		{
			Description: "valid definition with map schema",
			Schema:      map[string]interface{}{"$ref": "http://json-schema.org/draft-07/schema#"},
			Tests: []jsonSchemaTestCase{
				bsonTestCase(
					"valid definition schema",
					map[string]interface{}{
						"definitions": map[string]interface{}{
							"foo": map[string]interface{}{"bsonType": TYPE_INT32},
						},
					},
					true,
				),
				bsonTestCase(
					"valid definition schema with bson",
					bson.D{{"definitions", bson.D{{"foo", bson.D{{"bsonType", TYPE_INT32}}}}}},
					true,
				),
			},
		},
		{
			Description: "valid definition with map schema",
			Schema:      bson.D{{"$ref", "http://json-schema.org/draft-07/schema#"}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase(
					"valid definition schema",
					map[string]interface{}{
						"definitions": map[string]interface{}{
							"foo": map[string]interface{}{"bsonType": TYPE_INT32},
						},
					},
					true,
				),
				bsonTestCase(
					"valid definition schema with bson",
					bson.D{{"definitions", bson.D{{"foo", bson.D{{"bsonType", TYPE_INT32}}}}}},
					true,
				),
			},
		},
		{
			Description: "propertyNames with boolean schema false",
			Schema:      map[string]interface{}{"propertyNames": false},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("object with any properties is invalid", map[string]interface{}{"foo": 1}, false),
				bsonTestCase("object with any properties is invalid", map[string]interface{}{}, true),
				bsonTestCase("object with any properties is invalid as bson", bson.D{{"foo", 1}}, false),
				bsonTestCase("object with any properties is invalid as bson", bson.D{}, true),
			},
		},
		{
			Description: "propertyNames with boolean bson schema false",
			Schema:      bson.D{{"propertyNames", false}},
			Tests: []jsonSchemaTestCase{
				bsonTestCase("object with any properties is invalid", map[string]interface{}{"foo": 1}, false),
				bsonTestCase("object with any properties is invalid", map[string]interface{}{}, true),
				bsonTestCase("object with any properties is invalid as bson", bson.D{{"foo", 1}}, false),
				bsonTestCase("object with any properties is invalid as bson", bson.D{}, true),
			},
		},
	}
}
