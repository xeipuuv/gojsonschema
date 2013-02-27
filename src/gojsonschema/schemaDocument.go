// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"gojsonreference"
	"reflect"
)

func NewJsonSchemaDocument(documentReferenceString string) (*JsonSchemaDocument, error) {

	var err error

	d := JsonSchemaDocument{}
	d.documentReference, err = gojsonreference.NewJsonReference(documentReferenceString)

	document, err := GetHttpJson(documentReferenceString)
	if err != nil {
		return nil, err
	}

	err = d.parse(document)
	return &d, err
}

type JsonSchemaDocument struct {
	documentReference gojsonreference.JsonReference
	rootSchema        *JsonSchema
}

func (d *JsonSchemaDocument) parse(document interface{}) error {
	d.rootSchema = &JsonSchema{}
	return d.parseSchema(document, d.rootSchema)
}

func (d *JsonSchemaDocument) parseSchema(documentNode interface{}, currentSchema *JsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New("Schema must be an object")
	}

	m := documentNode.(map[string]interface{})

	if currentSchema == d.rootSchema {
		if !existsMapKey(m, "$schema") {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_IS_REQUIRED, "$schema"))
		}
		if !isKind(m["$schema"], reflect.String) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_MUST_BE_OF_TYPE, "$schema", "string"))
		}
		schemaRef := m["$schema"].(string)
		schemaReference, err := gojsonreference.NewJsonReference(schemaRef)
		currentSchema.schema = &schemaReference
		if err != nil {
			return err
		}
		
		currentSchema.ref = &d.documentReference
	}

	// ref
	if existsMapKey(m, "$ref") && !isKind(m["$ref"], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_MUST_BE_OF_TYPE, "$ref", "string"))
	}
	if k, ok := m["$ref"].(string); ok {
		jsonReference, err := gojsonreference.NewJsonReference(k)
		if err != nil {
			return err
		}
		inheritedReference, err := gojsonreference.Inherits(*currentSchema.ref,jsonReference) 
		if err != nil {
			return err
		}
		currentSchema.ref = inheritedReference
		httpDocumentNode, err := GetHttpJson(currentSchema.ref.String())
		if err != nil {
			return err
		}
		if !isKind(httpDocumentNode, reflect.Map) {
			return errors.New("Schema must be an object")
		}
		m = httpDocumentNode.(map[string]interface{})
	}

	// id
	if existsMapKey(m, "id") && !isKind(m["id"], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_MUST_BE_OF_TYPE, "id", "string"))
	}
	if k, ok := m["id"].(string); ok {
		currentSchema.id = &k
	}

	// title
	if existsMapKey(m, "title") && !isKind(m["title"], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_MUST_BE_OF_TYPE, "title", "string"))
	}
	if k, ok := m["title"].(string); ok {
		currentSchema.title = &k
	}

	// description
	if existsMapKey(m, "description") && !isKind(m["description"], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_MUST_BE_OF_TYPE, "description", "string"))
	}
	if k, ok := m["description"].(string); ok {
		currentSchema.description = &k
	}

	// properties
	/*	if !existsMapKey(m, "properties") {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_IS_REQUIRED, "properties"))
		}
	*/
	for k := range m {
		if k == "properties" {
			err := d.parseProperties(m[k], currentSchema)
			if err != nil {
				return err
			}
		}
	}

	// items
	/*	if !existsMapKey(m, "items") {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_IS_REQUIRED, "items"))
		}
	*/
	for k := range m {
		if k == "items" {
			newSchema := &JsonSchema{ parent : currentSchema }			
			currentSchema.AddPropertiesChild(newSchema)
			err := d.parseSchema(m[k], newSchema)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseProperties(documentNode interface{}, currentSchema *JsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New("Properties must be an object")
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		schemaProperty := k
		newSchema := &JsonSchema{property: &schemaProperty, parent : currentSchema, ref: currentSchema.ref}
		currentSchema.AddPropertiesChild(newSchema)
		err := d.parseSchema(m[k], newSchema)
		if err != nil {
			return err
		}
	}

	return nil
}
