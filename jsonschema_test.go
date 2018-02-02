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
	"time"
)

type jsonSchemaTest struct {
	Description string               `json:"description"`
	Disabled    bool                 `json:"disabled"`
	Schema      interface{}          `json:"schema"`
	Tests       []jsonSchemaTestCase `json:"tests"`
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

	testswd := filepath.Join(wd, "/testdata/draft4")

	go func() {
		err := http.ListenAndServe(":1234", http.FileServer(http.Dir(filepath.Join(wd, "/testdata/remotes/"))))
		if err != nil {
			panic(err.Error())
		}
	}()

	// time.Sleep(100 * time.Second)
	time.Sleep(time.Second)

	files, err := ioutil.ReadDir(testswd)
	if err != nil {
		panic(err.Error())
	}

	for _, fileInfo := range files {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".json") {
			filepath := filepath.Join(testswd, fileInfo.Name())

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
					// fmt.Println(test.Schema)
					t.Errorf("Error (%s)\n", err.Error())
				}
				for _, testCase := range test.Tests {
					testDataLoader := NewRawLoader(testCase.Data)
					result, err := testSchema.Validate(testDataLoader)
					if err != nil {
						t.Errorf("Error (%s)\n", err.Error())
					}
					if result.Valid() != testCase.Valid {
						fmt.Println("OOH MISMATCH")
						t.Errorf("Test failed : %s\n%s\n%s\nexpects: %t, given %t\n%s\n", file.Name(), test.Description, testCase.Description, testCase.Valid, result.Valid(), testCase.Data)

					}
				}
			}
		}
	}
}
