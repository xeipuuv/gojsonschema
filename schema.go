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
// description      Defines Schema, the main entry to every subSchema.
//                  Contains the parsing logic and error checking.
//
// created          26-02-2013

package gojsonschema

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"github.com/xeipuuv/gojsonreference"
	"reflect"
	"regexp"
)

func NewSchema(l JSONLoader) (*Schema, error) {
	return l.loadSchema()
}

type Schema struct {
	documentReference gojsonreference.JsonReference
	rootSchema        *subSchema
	pool              *schemaPool
	referencePool     *schemaReferencePool
}

func (d *Schema) parse(document interface{}) error {
	d.rootSchema = &subSchema{property: STRING_ROOT_SCHEMA_PROPERTY}
	return d.parseSchema(document, d.rootSchema)
}

func (d *Schema) SetRootSchemaName(name string) {
	d.rootSchema.property = name
}

// Parses a subSchema
//
// Pretty long function ( sorry :) )... but pretty straight forward, repetitive and boring
// Not much magic involved here, most of the job is to validate the key names and their values,
// then the values are copied into subSchema struct
//
func (d *Schema) parseSchema(documentNode interface{}, currentSchema *subSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, TYPE_OBJECT))
	}

	m := documentNode.(map[string]interface{})

	if currentSchema == d.rootSchema {
		currentSchema.ref = &d.documentReference
	}

	// $subSchema
	if existsMapKey(m, KEY_SCHEMA) {
		if !isKind(m[KEY_SCHEMA], reflect.String) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_SCHEMA, TYPE_STRING))
		}
		schemaRef := m[KEY_SCHEMA].(string)
		schemaReference, err := gojsonreference.NewJsonReference(schemaRef)
		currentSchema.subSchema = &schemaReference
		if err != nil {
			return err
		}
	}

	// $ref
	if existsMapKey(m, KEY_REF) && !isKind(m[KEY_REF], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_REF, TYPE_STRING))
	}
	if k, ok := m[KEY_REF].(string); ok {

		if sch, ok := d.referencePool.Get(currentSchema.ref.String() + k); ok {

			currentSchema.refSchema = sch

		} else {

			var err error
			err = d.parseReference(documentNode, currentSchema, k)
			if err != nil {
				return err
			}

			return nil
		}
	}

	// definitions
	if existsMapKey(m, KEY_DEFINITIONS) {
		if isKind(m[KEY_DEFINITIONS], reflect.Map) {
			currentSchema.definitions = make(map[string]*subSchema)
			for dk, dv := range m[KEY_DEFINITIONS].(map[string]interface{}) {
				if isKind(dv, reflect.Map) {
					newSchema := &subSchema{property: KEY_DEFINITIONS, parent: currentSchema, ref: currentSchema.ref}
					currentSchema.definitions[dk] = newSchema
					err := d.parseSchema(dv, newSchema)
					if err != nil {
						return errors.New(err.Error())
					}
				} else {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEFINITIONS, STRING_ARRAY_OF_SCHEMAS))
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEFINITIONS, STRING_ARRAY_OF_SCHEMAS))
		}

	}

	// id
	if existsMapKey(m, KEY_ID) && !isKind(m[KEY_ID], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ID, TYPE_STRING))
	}
	if k, ok := m[KEY_ID].(string); ok {
		currentSchema.id = &k
	}

	// title
	if existsMapKey(m, KEY_TITLE) && !isKind(m[KEY_TITLE], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TITLE, TYPE_STRING))
	}
	if k, ok := m[KEY_TITLE].(string); ok {
		currentSchema.title = &k
	}

	// description
	if existsMapKey(m, KEY_DESCRIPTION) && !isKind(m[KEY_DESCRIPTION], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DESCRIPTION, TYPE_STRING))
	}
	if k, ok := m[KEY_DESCRIPTION].(string); ok {
		currentSchema.description = &k
	}

	// type
	if existsMapKey(m, KEY_TYPE) {
		if isKind(m[KEY_TYPE], reflect.String) {
			if k, ok := m[KEY_TYPE].(string); ok {
				err := currentSchema.types.Add(k)
				if err != nil {
					return err
				}
			}
		} else {
			if isKind(m[KEY_TYPE], reflect.Slice) {
				arrayOfTypes := m[KEY_TYPE].([]interface{})
				for _, typeInArray := range arrayOfTypes {
					if reflect.ValueOf(typeInArray).Kind() != reflect.String {
						return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, TYPE_STRING+"/"+STRING_ARRAY_OF_STRINGS))
					} else {
						currentSchema.types.Add(typeInArray.(string))
					}
				}

			} else {
				return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, TYPE_STRING+"/"+STRING_ARRAY_OF_STRINGS))
			}
		}
	}

	// properties
	if existsMapKey(m, KEY_PROPERTIES) {
		err := d.parseProperties(m[KEY_PROPERTIES], currentSchema)
		if err != nil {
			return err
		}
	}

	// additionalProperties
	if existsMapKey(m, KEY_ADDITIONAL_PROPERTIES) {
		if isKind(m[KEY_ADDITIONAL_PROPERTIES], reflect.Bool) {
			currentSchema.additionalProperties = m[KEY_ADDITIONAL_PROPERTIES].(bool)
		} else if isKind(m[KEY_ADDITIONAL_PROPERTIES], reflect.Map) {
			newSchema := &subSchema{property: KEY_ADDITIONAL_PROPERTIES, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.additionalProperties = newSchema
			err := d.parseSchema(m[KEY_ADDITIONAL_PROPERTIES], newSchema)
			if err != nil {
				return errors.New(err.Error())
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ADDITIONAL_PROPERTIES, TYPE_BOOLEAN+"/"+STRING_SCHEMA))
		}
	}

	// patternProperties
	if existsMapKey(m, KEY_PATTERN_PROPERTIES) {
		if isKind(m[KEY_PATTERN_PROPERTIES], reflect.Map) {
			patternPropertiesMap := m[KEY_PATTERN_PROPERTIES].(map[string]interface{})
			if len(patternPropertiesMap) > 0 {
				currentSchema.patternProperties = make(map[string]*subSchema)
				for k, v := range patternPropertiesMap {
					_, err := regexp.MatchString(k, "")
					if err != nil {
						return errors.New(fmt.Sprintf(ERROR_MESSAGE_INVALID_REGEX_PATTERN, k))
					}
					newSchema := &subSchema{property: k, parent: currentSchema, ref: currentSchema.ref}
					err = d.parseSchema(v, newSchema)
					if err != nil {
						return errors.New(err.Error())
					}
					currentSchema.patternProperties[k] = newSchema
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_PATTERN_PROPERTIES, STRING_SCHEMA))
		}
	}

	// dependencies
	if existsMapKey(m, KEY_DEPENDENCIES) {
		err := d.parseDependencies(m[KEY_DEPENDENCIES], currentSchema)
		if err != nil {
			return err
		}
	}

	// items
	if existsMapKey(m, KEY_ITEMS) {
		if isKind(m[KEY_ITEMS], reflect.Slice) {
			for _, itemElement := range m[KEY_ITEMS].([]interface{}) {
				if isKind(itemElement, reflect.Map) {
					newSchema := &subSchema{parent: currentSchema, property: KEY_ITEMS}
					newSchema.ref = currentSchema.ref
					currentSchema.AddItemsChild(newSchema)
					err := d.parseSchema(itemElement, newSchema)
					if err != nil {
						return err
					}
				} else {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ITEMS, STRING_SCHEMA+"/"+STRING_ARRAY_OF_SCHEMAS))
				}
				currentSchema.itemsChildrenIsSingleSchema = false
			}
		} else if isKind(m[KEY_ITEMS], reflect.Map) {
			newSchema := &subSchema{parent: currentSchema, property: KEY_ITEMS}
			newSchema.ref = currentSchema.ref
			currentSchema.AddItemsChild(newSchema)
			err := d.parseSchema(m[KEY_ITEMS], newSchema)
			if err != nil {
				return err
			}
			currentSchema.itemsChildrenIsSingleSchema = true
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ITEMS, STRING_SCHEMA+"/"+STRING_ARRAY_OF_SCHEMAS))
		}
	}

	// additionalItems
	if existsMapKey(m, KEY_ADDITIONAL_ITEMS) {
		if isKind(m[KEY_ADDITIONAL_ITEMS], reflect.Bool) {
			currentSchema.additionalItems = m[KEY_ADDITIONAL_ITEMS].(bool)
		} else if isKind(m[KEY_ADDITIONAL_ITEMS], reflect.Map) {
			newSchema := &subSchema{property: KEY_ADDITIONAL_ITEMS, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.additionalItems = newSchema
			err := d.parseSchema(m[KEY_ADDITIONAL_ITEMS], newSchema)
			if err != nil {
				return errors.New(err.Error())
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ADDITIONAL_ITEMS, TYPE_BOOLEAN+"/"+STRING_SCHEMA))
		}
	}

	// validation : number / integer

	if existsMapKey(m, KEY_MULTIPLE_OF) {
		multipleOfValue := mustBeNumber(m[KEY_MULTIPLE_OF])
		if multipleOfValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_MULTIPLE_OF, STRING_NUMBER))
		}
		if *multipleOfValue <= 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_STRICTLY_GREATER_THAN_0, KEY_MULTIPLE_OF))
		}
		currentSchema.multipleOf = multipleOfValue
	}

	if existsMapKey(m, KEY_MINIMUM) {
		minimumValue := mustBeNumber(m[KEY_MINIMUM])
		if minimumValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_MINIMUM, STRING_NUMBER))
		}
		currentSchema.minimum = minimumValue
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MINIMUM) {
		if isKind(m[KEY_EXCLUSIVE_MINIMUM], reflect.Bool) {
			if currentSchema.minimum == nil {
				return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_CANNOT_BE_USED_WITHOUT_Y, KEY_EXCLUSIVE_MINIMUM, KEY_MINIMUM))
			}
			exclusiveMinimumValue := m[KEY_EXCLUSIVE_MINIMUM].(bool)
			currentSchema.exclusiveMinimum = exclusiveMinimumValue
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_EXCLUSIVE_MINIMUM, TYPE_BOOLEAN))
		}
	}

	if existsMapKey(m, KEY_MAXIMUM) {
		maximumValue := mustBeNumber(m[KEY_MAXIMUM])
		if maximumValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_MAXIMUM, STRING_NUMBER))
		}
		currentSchema.maximum = maximumValue
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MAXIMUM) {
		if isKind(m[KEY_EXCLUSIVE_MAXIMUM], reflect.Bool) {
			if currentSchema.maximum == nil {
				return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_CANNOT_BE_USED_WITHOUT_Y, KEY_EXCLUSIVE_MAXIMUM, KEY_MAXIMUM))
			}
			exclusiveMaximumValue := m[KEY_EXCLUSIVE_MAXIMUM].(bool)
			currentSchema.exclusiveMaximum = exclusiveMaximumValue
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_EXCLUSIVE_MAXIMUM, STRING_NUMBER))
		}
	}

	if currentSchema.minimum != nil && currentSchema.maximum != nil {
		if *currentSchema.minimum > *currentSchema.maximum {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_CANNOT_BE_GREATER_THAN_Y, KEY_MINIMUM, KEY_MAXIMUM))
		}
	}

	// validation : string

	if existsMapKey(m, KEY_MIN_LENGTH) {
		minLengthIntegerValue := mustBeInteger(m[KEY_MIN_LENGTH])
		if minLengthIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MIN_LENGTH, TYPE_INTEGER))
		}
		if *minLengthIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MIN_LENGTH))
		}
		currentSchema.minLength = minLengthIntegerValue
	}

	if existsMapKey(m, KEY_MAX_LENGTH) {
		maxLengthIntegerValue := mustBeInteger(m[KEY_MAX_LENGTH])
		if maxLengthIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MAX_LENGTH, TYPE_INTEGER))
		}
		if *maxLengthIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MAX_LENGTH))
		}
		currentSchema.maxLength = maxLengthIntegerValue
	}

	if currentSchema.minLength != nil && currentSchema.maxLength != nil {
		if *currentSchema.minLength > *currentSchema.maxLength {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_CANNOT_BE_GREATER_THAN_Y, KEY_MIN_LENGTH, KEY_MAX_LENGTH))
		}
	}

	if existsMapKey(m, KEY_PATTERN) {
		if isKind(m[KEY_PATTERN], reflect.String) {
			regexpObject, err := regexp.Compile(m[KEY_PATTERN].(string))
			if err != nil {
				return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_VALID_REGEX, KEY_PATTERN))
			}
			currentSchema.pattern = regexpObject
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_PATTERN, TYPE_STRING))
		}
	}

	// validation : object

	if existsMapKey(m, KEY_MIN_PROPERTIES) {
		minPropertiesIntegerValue := mustBeInteger(m[KEY_MIN_PROPERTIES])
		if minPropertiesIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MIN_PROPERTIES, TYPE_INTEGER))
		}
		if *minPropertiesIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MIN_PROPERTIES))
		}
		currentSchema.minProperties = minPropertiesIntegerValue
	}

	if existsMapKey(m, KEY_MAX_PROPERTIES) {
		maxPropertiesIntegerValue := mustBeInteger(m[KEY_MAX_PROPERTIES])
		if maxPropertiesIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MAX_PROPERTIES, TYPE_INTEGER))
		}
		if *maxPropertiesIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MAX_PROPERTIES))
		}
		currentSchema.maxProperties = maxPropertiesIntegerValue
	}

	if currentSchema.minProperties != nil && currentSchema.maxProperties != nil {
		if *currentSchema.minProperties > *currentSchema.maxProperties {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_CANNOT_BE_GREATER_THAN_Y, KEY_MIN_PROPERTIES, KEY_MAX_PROPERTIES))
		}
	}

	if existsMapKey(m, KEY_REQUIRED) {
		if isKind(m[KEY_REQUIRED], reflect.Slice) {
			requiredValues := m[KEY_REQUIRED].([]interface{})
			for _, requiredValue := range requiredValues {
				if isKind(requiredValue, reflect.String) {
					err := currentSchema.AddRequired(requiredValue.(string))
					if err != nil {
						return err
					}
				} else {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_ITEMS_MUST_BE_TYPE_Y, KEY_REQUIRED, TYPE_STRING))
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_REQUIRED, TYPE_ARRAY))
		}
	}

	// validation : array

	if existsMapKey(m, KEY_MIN_ITEMS) {
		minItemsIntegerValue := mustBeInteger(m[KEY_MIN_ITEMS])
		if minItemsIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MIN_ITEMS, TYPE_INTEGER))
		}
		if *minItemsIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MIN_ITEMS))
		}
		currentSchema.minItems = minItemsIntegerValue
	}

	if existsMapKey(m, KEY_MAX_ITEMS) {
		maxItemsIntegerValue := mustBeInteger(m[KEY_MAX_ITEMS])
		if maxItemsIntegerValue == nil {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_MAX_ITEMS, TYPE_INTEGER))
		}
		if *maxItemsIntegerValue < 0 {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_GREATER_OR_TO_0, KEY_MAX_ITEMS))
		}
		currentSchema.maxItems = maxItemsIntegerValue
	}

	if existsMapKey(m, KEY_UNIQUE_ITEMS) {
		if isKind(m[KEY_UNIQUE_ITEMS], reflect.Bool) {
			currentSchema.uniqueItems = m[KEY_UNIQUE_ITEMS].(bool)
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_A_Y, KEY_UNIQUE_ITEMS, TYPE_BOOLEAN))
		}
	}

	// validation : all

	if existsMapKey(m, KEY_ENUM) {
		if isKind(m[KEY_ENUM], reflect.Slice) {
			for _, v := range m[KEY_ENUM].([]interface{}) {
				err := currentSchema.AddEnum(v)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_ENUM, TYPE_ARRAY))
		}
	}

	// validation : subSchema

	if existsMapKey(m, KEY_ONE_OF) {
		if isKind(m[KEY_ONE_OF], reflect.Slice) {
			for _, v := range m[KEY_ONE_OF].([]interface{}) {
				newSchema := &subSchema{property: KEY_ONE_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddOneOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_ONE_OF, TYPE_ARRAY))
		}
	}

	if existsMapKey(m, KEY_ANY_OF) {
		if isKind(m[KEY_ANY_OF], reflect.Slice) {
			for _, v := range m[KEY_ANY_OF].([]interface{}) {
				newSchema := &subSchema{property: KEY_ANY_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddAnyOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_ANY_OF, TYPE_ARRAY))
		}
	}

	if existsMapKey(m, KEY_ALL_OF) {
		if isKind(m[KEY_ALL_OF], reflect.Slice) {
			for _, v := range m[KEY_ALL_OF].([]interface{}) {
				newSchema := &subSchema{property: KEY_ALL_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddAllOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_ANY_OF, TYPE_ARRAY))
		}
	}

	if existsMapKey(m, KEY_NOT) {
		if isKind(m[KEY_NOT], reflect.Map) {
			newSchema := &subSchema{property: KEY_NOT, parent: currentSchema, ref: currentSchema.ref}
			currentSchema.SetNot(newSchema)
			err := d.parseSchema(m[KEY_NOT], newSchema)
			if err != nil {
				return err
			}
		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_Y, KEY_NOT, TYPE_OBJECT))
		}
	}

	return nil
}

func (d *Schema) parseReference(documentNode interface{}, currentSchema *subSchema, reference string) (e error) {

	var err error

	jsonReference, err := gojsonreference.NewJsonReference(reference)
	if err != nil {
		return err
	}

	standaloneDocument := d.pool.GetStandaloneDocument()

	if jsonReference.HasFullUrl {
		currentSchema.ref = &jsonReference
	} else {
		inheritedReference, err := currentSchema.ref.Inherits(jsonReference)
		if err != nil {
			return err
		}
		currentSchema.ref = inheritedReference
	}

	jsonPointer := currentSchema.ref.GetPointer()

	var refdDocumentNode interface{}

	if standaloneDocument != nil {

		var err error
		refdDocumentNode, _, err = jsonPointer.Get(standaloneDocument)
		if err != nil {
			return err
		}

	} else {

		var err error
		dsp, err := d.pool.GetDocument(*currentSchema.ref)
		if err != nil {
			return err
		}

		refdDocumentNode, _, err = jsonPointer.Get(dsp.Document)
		if err != nil {
			return err
		}

	}

	if !isKind(refdDocumentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, TYPE_OBJECT))
	}

	// returns the loaded referenced subSchema for the caller to update its current subSchema
	newSchemaDocument := refdDocumentNode.(map[string]interface{})

	newSchema := &subSchema{property: KEY_REF, parent: currentSchema, ref: currentSchema.ref}
	d.referencePool.Add(currentSchema.ref.String()+reference, newSchema)

	err = d.parseSchema(newSchemaDocument, newSchema)
	if err != nil {
		return err
	}

	currentSchema.refSchema = newSchema

	return nil

}

func (d *Schema) parseProperties(documentNode interface{}, currentSchema *subSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_PROPERTIES, TYPE_OBJECT))
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		schemaProperty := k
		newSchema := &subSchema{property: schemaProperty, parent: currentSchema, ref: currentSchema.ref}
		currentSchema.AddPropertiesChild(newSchema)
		err := d.parseSchema(m[k], newSchema)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *Schema) parseDependencies(documentNode interface{}, currentSchema *subSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DEPENDENCIES, TYPE_OBJECT))
	}

	m := documentNode.(map[string]interface{})
	currentSchema.dependencies = make(map[string]interface{})

	for k := range m {
		switch reflect.ValueOf(m[k]).Kind() {

		case reflect.Slice:
			values := m[k].([]interface{})
			var valuesToRegister []string

			for _, value := range values {
				if !isKind(value, reflect.String) {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_DEPENDENCY, STRING_SCHEMA_OR_ARRAY_OF_STRINGS))
				} else {
					valuesToRegister = append(valuesToRegister, value.(string))
				}
				currentSchema.dependencies[k] = valuesToRegister
			}

		case reflect.Map:
			depSchema := &subSchema{property: k, parent: currentSchema, ref: currentSchema.ref}
			err := d.parseSchema(m[k], depSchema)
			if err != nil {
				return err
			}
			currentSchema.dependencies[k] = depSchema

		default:
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_DEPENDENCY, STRING_SCHEMA_OR_ARRAY_OF_STRINGS))
		}

	}

	return nil
}
