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
// description		Different strategies to load JSON files.
// 					Includes References (file and HTTP), JSON strings and Go types.
//
// created          01-02-2015

package gojsonschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/xeipuuv/gojsonreference"
)

var osFS = osFileSystem(os.Open)

// JSON loader interface

type JSONLoader interface {
	JsonSource() interface{}
	LoadJSON() (interface{}, error)
	JsonReference() (gojsonreference.JsonReference, error)
	LoaderFactory() JSONLoaderFactory
}

type JSONLoaderFactory interface {
	New(source string) JSONLoader
}

type DefaultJSONLoaderFactory struct {
}

type FileSystemJSONLoaderFactory struct {
	fs http.FileSystem
}

func (d DefaultJSONLoaderFactory) New(source string) JSONLoader {
	return &jsonReferenceLoader{
		fs:     osFS,
		source: source,
	}
}

func (f FileSystemJSONLoaderFactory) New(source string) JSONLoader {
	return &jsonReferenceLoader{
		fs:     f.fs,
		source: source,
	}
}

// osFileSystem is a functional wrapper for os.Open that implements http.FileSystem.
type osFileSystem func(string) (*os.File, error)

func (o osFileSystem) Open(name string) (http.File, error) {
	return o(name)
}

// JSON Reference loader
// references are used to load JSONs from files and HTTP

type jsonReferenceLoader struct {
	fs     http.FileSystem
	source string
}

func (l *jsonReferenceLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonReferenceLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference(l.JsonSource().(string))
}

func (l *jsonReferenceLoader) LoaderFactory() JSONLoaderFactory {
	return &FileSystemJSONLoaderFactory{
		fs: l.fs,
	}
}

// NewReferenceLoader returns a JSON reference loader using the given source and the local OS file system.
func NewReferenceLoader(source string) JSONLoader {
	return &jsonReferenceLoader{
		fs:     osFS,
		source: source,
	}
}

// NewReferenceLoaderFileSystem returns a JSON reference loader using the given source and file system.
func NewReferenceLoaderFileSystem(source string, fs http.FileSystem) JSONLoader {
	return &jsonReferenceLoader{
		fs:     fs,
		source: source,
	}
}

func (l *jsonReferenceLoader) LoadJSON() (interface{}, error) {

	var err error

	reference, err := gojsonreference.NewJsonReference(l.JsonSource().(string))
	if err != nil {
		return nil, err
	}

	refToUrl := reference
	refToUrl.GetUrl().Fragment = ""

	var document interface{}

	if reference.HasFileScheme {

		filename := strings.TrimPrefix(refToUrl.String(), "file://")
		if runtime.GOOS == "windows" {
			// on Windows, a file URL may have an extra leading slash, use slashes
			// instead of backslashes, and have spaces escaped
			filename = strings.TrimPrefix(filename, "/")
			filename = filepath.FromSlash(filename)
		}

		document, err = l.loadFromFile(filename)
		if err != nil {
			return nil, err
		}

	} else {

		document, err = l.loadFromHTTP(refToUrl.String())
		if err != nil {
			return nil, err
		}

	}

	return document, nil

}

func (l *jsonReferenceLoader) loadFromHTTP(address string) (interface{}, error) {

	switch address {
	case "http://json-schema.org/draft-04/schema":
		return decodeJsonUsingNumber(strings.NewReader(`{"id":"http://json-schema.org/draft-04/schema#","$schema":"http://json-schema.org/draft-04/schema#","description":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"positiveInteger":{"type":"integer","minimum":0},"positiveIntegerDefault0":{"allOf":[{"$ref":"#/definitions/positiveInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"minItems":1,"uniqueItems":true}},"type":"object","properties":{"id":{"type":"string"},"$schema":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"multipleOf":{"type":"number","minimum":0,"exclusiveMinimum":true},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"boolean","default":false},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"boolean","default":false},"maxLength":{"$ref":"#/definitions/positiveInteger"},"minLength":{"$ref":"#/definitions/positiveIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/positiveInteger"},"minItems":{"$ref":"#/definitions/positiveIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"maxProperties":{"$ref":"#/definitions/positiveInteger"},"minProperties":{"$ref":"#/definitions/positiveIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"anyOf":[{"type":"boolean"},{"$ref":"#"}],"default":{}},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"dependencies":{"exclusiveMaximum":["maximum"],"exclusiveMinimum":["minimum"]},"default":{}}`))
	case "http://json-schema.org/draft-06/schema":
		return decodeJsonUsingNumber(strings.NewReader(`{"$schema":"http://json-schema.org/draft-06/schema#","$id":"http://json-schema.org/draft-06/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"title":{"type":"string"},"description":{"type":"string"},"default":{},"examples":{"type":"array","items":{}},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":{}},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":{},"enum":{"type":"array","minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":{}}`))
	case "http://json-schema.org/draft-07/schema":
		return decodeJsonUsingNumber(strings.NewReader(`{"$schema":"http://json-schema.org/draft-07/schema#","$id":"http://json-schema.org/draft-07/schema#","title":"Core schema meta-schema","definitions":{"schemaArray":{"type":"array","minItems":1,"items":{"$ref":"#"}},"nonNegativeInteger":{"type":"integer","minimum":0},"nonNegativeIntegerDefault0":{"allOf":[{"$ref":"#/definitions/nonNegativeInteger"},{"default":0}]},"simpleTypes":{"enum":["array","boolean","integer","null","number","object","string"]},"stringArray":{"type":"array","items":{"type":"string"},"uniqueItems":true,"default":[]}},"type":["object","boolean"],"properties":{"$id":{"type":"string","format":"uri-reference"},"$schema":{"type":"string","format":"uri"},"$ref":{"type":"string","format":"uri-reference"},"$comment":{"type":"string"},"title":{"type":"string"},"description":{"type":"string"},"default":true,"readOnly":{"type":"boolean","default":false},"examples":{"type":"array","items":true},"multipleOf":{"type":"number","exclusiveMinimum":0},"maximum":{"type":"number"},"exclusiveMaximum":{"type":"number"},"minimum":{"type":"number"},"exclusiveMinimum":{"type":"number"},"maxLength":{"$ref":"#/definitions/nonNegativeInteger"},"minLength":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"pattern":{"type":"string","format":"regex"},"additionalItems":{"$ref":"#"},"items":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/schemaArray"}],"default":true},"maxItems":{"$ref":"#/definitions/nonNegativeInteger"},"minItems":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"uniqueItems":{"type":"boolean","default":false},"contains":{"$ref":"#"},"maxProperties":{"$ref":"#/definitions/nonNegativeInteger"},"minProperties":{"$ref":"#/definitions/nonNegativeIntegerDefault0"},"required":{"$ref":"#/definitions/stringArray"},"additionalProperties":{"$ref":"#"},"definitions":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"properties":{"type":"object","additionalProperties":{"$ref":"#"},"default":{}},"patternProperties":{"type":"object","additionalProperties":{"$ref":"#"},"propertyNames":{"format":"regex"},"default":{}},"dependencies":{"type":"object","additionalProperties":{"anyOf":[{"$ref":"#"},{"$ref":"#/definitions/stringArray"}]}},"propertyNames":{"$ref":"#"},"const":true,"enum":{"type":"array","items":true,"minItems":1,"uniqueItems":true},"type":{"anyOf":[{"$ref":"#/definitions/simpleTypes"},{"type":"array","items":{"$ref":"#/definitions/simpleTypes"},"minItems":1,"uniqueItems":true}]},"format":{"type":"string"},"contentMediaType":{"type":"string"},"contentEncoding":{"type":"string"},"if":{"$ref":"#"},"then":{"$ref":"#"},"else":{"$ref":"#"},"allOf":{"$ref":"#/definitions/schemaArray"},"anyOf":{"$ref":"#/definitions/schemaArray"},"oneOf":{"$ref":"#/definitions/schemaArray"},"not":{"$ref":"#"}},"default":true}`))
	default:

		resp, err := http.Get(address)
		if err != nil {
			return nil, err
		}

		// must return HTTP Status 200 OK
		if resp.StatusCode != http.StatusOK {
			return nil, errors.New(formatErrorDescription(Locale.HttpBadStatus(), ErrorDetails{"status": resp.Status}))
		}

		bodyBuff, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return decodeJsonUsingNumber(bytes.NewReader(bodyBuff))
	}
}

func (l *jsonReferenceLoader) loadFromFile(path string) (interface{}, error) {
	f, err := l.fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bodyBuff, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return decodeJsonUsingNumber(bytes.NewReader(bodyBuff))

}

// JSON string loader

type jsonStringLoader struct {
	source string
}

func (l *jsonStringLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonStringLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonStringLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

func NewStringLoader(source string) JSONLoader {
	return &jsonStringLoader{source: source}
}

func (l *jsonStringLoader) LoadJSON() (interface{}, error) {

	return decodeJsonUsingNumber(strings.NewReader(l.JsonSource().(string)))

}

// JSON bytes loader

type jsonBytesLoader struct {
	source []byte
}

func (l *jsonBytesLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonBytesLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonBytesLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

func NewBytesLoader(source []byte) JSONLoader {
	return &jsonBytesLoader{source: source}
}

func (l *jsonBytesLoader) LoadJSON() (interface{}, error) {
	return decodeJsonUsingNumber(bytes.NewReader(l.JsonSource().([]byte)))
}

// JSON Go (types) loader
// used to load JSONs from the code as maps, interface{}, structs ...

type jsonGoLoader struct {
	source interface{}
}

func (l *jsonGoLoader) JsonSource() interface{} {
	return l.source
}

func (l *jsonGoLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonGoLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

func NewGoLoader(source interface{}) JSONLoader {
	return &jsonGoLoader{source: source}
}

func (l *jsonGoLoader) LoadJSON() (interface{}, error) {

	// convert it to a compliant JSON first to avoid types "mismatches"

	jsonBytes, err := json.Marshal(l.JsonSource())
	if err != nil {
		return nil, err
	}

	return decodeJsonUsingNumber(bytes.NewReader(jsonBytes))

}

type jsonIOLoader struct {
	buf *bytes.Buffer
}

func NewReaderLoader(source io.Reader) (JSONLoader, io.Reader) {
	buf := &bytes.Buffer{}
	return &jsonIOLoader{buf: buf}, io.TeeReader(source, buf)
}

func NewWriterLoader(source io.Writer) (JSONLoader, io.Writer) {
	buf := &bytes.Buffer{}
	return &jsonIOLoader{buf: buf}, io.MultiWriter(source, buf)
}

func (l *jsonIOLoader) JsonSource() interface{} {
	return l.buf.String()
}

func (l *jsonIOLoader) LoadJSON() (interface{}, error) {
	return decodeJsonUsingNumber(l.buf)
}

func (l *jsonIOLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}

func (l *jsonIOLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

// JSON raw loader
// In case the JSON is already marshalled to interface{} use this loader
// This is used for testing as otherwise there is no guarantee the JSON is marshalled
// "properly" by using https://golang.org/pkg/encoding/json/#Decoder.UseNumber
type jsonRawLoader struct {
	source interface{}
}

func NewRawLoader(source interface{}) *jsonRawLoader {
	return &jsonRawLoader{source: source}
}
func (l *jsonRawLoader) JsonSource() interface{} {
	return l.source
}
func (l *jsonRawLoader) LoadJSON() (interface{}, error) {
	return l.source, nil
}
func (l *jsonRawLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference("#")
}
func (l *jsonRawLoader) LoaderFactory() JSONLoaderFactory {
	return &DefaultJSONLoaderFactory{}
}

func decodeJsonUsingNumber(r io.Reader) (interface{}, error) {

	var document interface{}

	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	err := decoder.Decode(&document)
	if err != nil {
		return nil, err
	}

	return document, nil

}
