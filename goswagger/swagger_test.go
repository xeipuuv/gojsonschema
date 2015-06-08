package goswagger

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	tt "github.com/apcera/continuum/util/testtool"
	"github.com/apcera/gojsonschema"
)

type TestData struct {
	Operation                string
	Path                     string
	Body                     string
	expectedProcessingError  bool
	expectedValid            bool
	expectedValidationErrors int
}

func TestBasic(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"PUT", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c"}`, false, true, 0},
		{"POST", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, false, true, 0},
	}
	ProcessData(t, TestDataSet)
}

func TestMissingRequired(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"PUT", "/jobs/{uuid}", `{"name":"a", "fqn":"b"}`, false, false, 1},
		{"POST", "/jobs/{uuid}", `{"name":"a", "uuid":"c", "otherField":0}`, false, false, 1},
	}
	ProcessData(t, TestDataSet)
}

func TestInvalidDataType(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"PUT", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":3}`, false, false, 1},
		{"POST", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c", "ports":3}`, false, false, 1},
	}
	ProcessData(t, TestDataSet)
}
func TestInvalidDataValue(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"POST", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c", "ports":{"number":0}}`, false, true, 0},
		{"POST", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c", "ports":{"number":-1}}`, false, false, 1},
		{"POST", "/jobs/{uuid}", `{"name":"a", "fqn":"b", "uuid":"c", "ports":{"number":99999999}}`, false, false, 1},
	}
	ProcessData(t, TestDataSet)
}

// Currently, pattern matching is quite forgiving and allows any characters between the '/'.
func TestPathMatching(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"PUT", "/jobs/abc", `{"name":"a", "fqn":"b", "uuid":"c"}`, false, true, 0},
		{"POST", "/jobs/993-a93", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, false, true, 0},
		{"POST", "/jobs/993&*#@93", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, false, true, 0},
		{"POST", "/jobs/99\\3\"&*#@93", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, false, true, 0},
	}
	ProcessData(t, TestDataSet)
}
func TestInvalidPaths(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"POST", "/jobs/xyz/", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, true, true, 0},
		{"POST", "/jobs/xyz/a", `{"name":"a", "fqn":"b", "uuid":"c", "otherField":0}`, true, true, 0},
	}
	ProcessData(t, TestDataSet)
}

// This functionality is supported through gojsonschema but is not supported by Swagger specification.
// This is due to uncertainty on their part of how to implement some of their SwaggerUI when there is a
// range of conditions that are combined with 'AND' or 'OR' statements (allOf, anyOf, oneOf, etc).
func TestAllOf(t *testing.T) {
	testHelper := tt.StartTest(t)
	defer testHelper.FinishTest()

	TestDataSet := []TestData{
		{"POST", "/jobs/testAllOf", `{"name":"a", "fqn":"b", "uuid":"c", "ports":{"number":0}}`, false, true, 0},
		// Two errors since violation ports structure (missing) and also a general failure of 'allOf'
		{"POST", "/jobs/testAllOf", `{"name":"a", "fqn":"b", "uuid":"c"}`, false, false, 2},
	}
	ProcessData(t, TestDataSet)
}

func ProcessData(t *testing.T, ds []TestData) {
	for _, d := range ds {
		result, err := ValidateData(t, d.Operation, d.Path, d.Body)
		if d.expectedProcessingError {
			t.Logf("Processing error when validating: %s", err)
			tt.TestExpectError(t, err)
			continue
		} else {
			tt.TestExpectSuccess(t, err)
		}

		tt.TestNotEqual(t, nil, result)
		if len(result.Errors()) > 0 {
			t.Log("RESULT: The document is not valid. see errors :")
			for _, desc := range result.Errors() {
				t.Logf("- %s", desc)
			}
		}
		tt.TestEqual(t, result.Valid(), d.expectedValid)
		tt.TestEqual(t, len(result.Errors()), d.expectedValidationErrors)
	}
}

func ValidateData(t *testing.T, op string, path string, body string) (*gojsonschema.Result, error) {
	t.Logf("Validating test case %s %s %s", op, path, body)

	s, err := ioutil.ReadFile("testSwaggerSpec.json")
	if err != nil {
		t.Error("Error reading swagger spec: %s", err)
		return nil, err
	}
	swag, err := GetSwaggerSpecFromBytes(s)
	if err != nil {
		t.Errorf("Error while getting swagger spec: %s", err)
		return nil, err
	}

	bodyReader := strings.NewReader(body)
	req, err := http.NewRequest(op, path, bodyReader)
	if err != nil {
		t.Errorf("Error while creating HTTP request: %s", err)
		return nil, err
	}

	result, err := swag.ValidateHTTPRequest(req)
	return result, err
}
