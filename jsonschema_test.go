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
	"regexp"
	"strings"
	"testing"
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
	Description string                `json:"description"`
	Data        interface{}           `json:"data"`
	Valid       bool                  `json:"valid"`
	Errors      []jsonSchemaTestError `json:"errors"`
}

type jsonSchemaTestError struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Context     string `json:"context"` // .Context().String()
	Description string `json:"description"`
	Details     string `json:"details"` // fmt.Sprintf %v of .Details() map[string]string
	String      string `json:"string"`
}

func asTestError(re ResultError) jsonSchemaTestError {
	rei := jsonSchemaTestError{
		Field:       re.Field(),
		Type:        re.Type(),
		Context:     re.Context().String(),
		Description: re.Description(),
		Details:     fmt.Sprintf("%v", re.Details()),
		String:      re.String(),
	}
	return rei
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

		if test.Disabled {
			continue
		}

		t.Run(test.Description, func(t *testing.T) {

			sl := NewSchemaLoader()
			sl.Draft = draft
			sl.Validate = true
			testSchema, err := sl.Compile(NewRawLoader(test.Schema))
			if err != nil {
				t.Errorf("Error (%s)\n", err.Error())
			}

			for _, testCase := range test.Tests {
				t.Run(testCase.Description, func(t *testing.T) {

					result, err := testSchema.Validate(NewRawLoader(testCase.Data))
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

					// if validation failed, and that was expected, then test the result.Errors()
					if !result.Valid() && !testCase.Valid {
						tcerrs := testCase.Errors
						if tcerrs == nil {
							t.Log("'errors' not set in test case so no coverage")
							return
						}
						rerrs := result.Errors()
						if len(rerrs) != len(tcerrs) {
							t.Errorf("expected %d errors but got %d", len(tcerrs), len(rerrs))
						}
						for i, re := range result.Errors() {
							jRE, _ := marshalToJSONString(asTestError(re))
							jTC, _ := marshalToJSONString(tcerrs[i])
							if *jRE != *jTC {
								t.Errorf("result.Errors %d: expected\n%s got\n%s", i, *jRE, *jTC)
							}
						}
					}
				})
			}
		})
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
		err := http.ListenAndServe("localhost:1234", http.FileServer(http.Dir(filepath.Join(wd, "remotes"))))
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
			t.Run(dir.Name(), func(t *testing.T) {
				formatJSONFile := filepath.Join(wd, dir.Name(), "optional", "format.json")
				if _, err = os.Stat(formatJSONFile); err == nil {
					err = executeTests(t, formatJSONFile)
				} else {
					err = nil
				}

				if err != nil {
					t.Errorf("Error (%s)\n", err.Error())
				}

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
			})
		}
	}
}
