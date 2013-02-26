// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"reflect"
)

func NewJsonSchemaDocument(document interface{}) (JsonSchemaDocument, error) {
	var d JsonSchemaDocument
	err := d.parse(document)
	return d, err
}

type JsonSchemaDocument struct {
}

func (d *JsonSchemaDocument) parse(document interface{}) error {
	return d.parseNodeRecursive(document)
}

func (d *JsonSchemaDocument) parseNodeRecursive(documentNode interface{}) error {

	rValue := reflect.ValueOf(documentNode)
	rKind := rValue.Kind()

	fmt.Printf("Kind : %s\n", rKind)

	switch rKind {

	case reflect.Map:
		m := documentNode.(map[string]interface{})
		for k := range m {
			fmt.Printf("Key : %s\n", k)
			err := d.parseNodeRecursive(m[k])
			if err != nil {
				return err
			}
		}

	case reflect.String:
	case reflect.Bool:

	default:
		return errors.New(fmt.Sprintf("Unhandled kind : %s", rKind))
	}

	return nil
}
