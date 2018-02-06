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
	"strings"
	"testing"
)

type jsonSchemaTest struct {
	Description string `json:"description"`
	// Some tests do not pass yet, so some tests are manually edited to include
	// an extra attribute whether that specific test should be disabled and skipped
	Disabled bool                 `json:"disabled"`
	Schema   interface{}          `json:"schema"`
	Tests    []jsonSchemaTestCase `json:"tests"`
}
type jsonSchemaTestCase struct {
	Description string      `json:"description"`
	Data        interface{} `json:"data"`
	Valid       bool        `json:"valid"`
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

	testDirectories := []string{"draft4", "draft6", "draft7"}

	var files []string
	for _, testDirectory := range testDirectories {
		testFiles, err := ioutil.ReadDir(filepath.Join(wd, testDirectory))

		if err != nil {
			panic(err.Error())
		}

		for _, fileInfo := range testFiles {
			if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".json") {
				files = append(files, filepath.Join(wd, testDirectory, fileInfo.Name()))
			}
		}
	}

	for _, filepath := range files {

		file, err := os.Open(filepath)
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

		for _, test := range tests {

			if test.Disabled {
				continue
			}

			testSchemaLoader := NewRawLoader(test.Schema)
			testSchema, err := NewSchema(testSchemaLoader)

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
					schemaString, _ := marshalToJsonString(test.Schema)
					testCaseString, _ := marshalToJsonString(testCase.Data)

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
	}
}
