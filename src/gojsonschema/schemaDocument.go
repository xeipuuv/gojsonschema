// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"gojsonreference"
	"io/ioutil"
	"net/http"
	"reflect"
)

func NewJsonSchemaDocument(documentReferenceString string) (*JsonSchemaDocument, error) {

	var err error

	d := JsonSchemaDocument{}
	d.documentReference, err = gojsonreference.NewJsonReference(documentReferenceString)

	resp, err := http.Get(documentReferenceString)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Could not access schema " + resp.Status)
	}

	bodyBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var document interface{}
	err = json.Unmarshal(bodyBuff, &document)
	if err != nil {
		return nil, err
	}

	err = d.parse(document)
	return &d, err
}

type JsonSchemaDocument struct {
	documentReference gojsonreference.JsonReference
}

func (d *JsonSchemaDocument) parse(document interface{}) error {
	return d.parseSchema(document)
}

func (d *JsonSchemaDocument) parseSchema(documentNode interface{}) error {
	fmt.Printf("-Schema\n")

	rValue := reflect.ValueOf(documentNode)
	rKind := rValue.Kind()

	if rKind != reflect.Map {
		return errors.New("Schema must be an object")
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		if k == "properties" {
			err := d.parseProperties(m[k])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseProperties(documentNode interface{}) error {
	fmt.Printf("-Properties\n")

	rValue := reflect.ValueOf(documentNode)
	rKind := rValue.Kind()

	if rKind != reflect.Map {
		return errors.New("Properties must be an object")
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		err := d.parseSchema(m[k])
		if err != nil {
			return err
		}
	}

	return nil
}
