// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      26-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"gojsonreference"
	"reflect"
	"regexp"
)

func NewJsonSchemaDocument(documentReferenceString string) (*JsonSchemaDocument, error) {

	var err error

	d := JsonSchemaDocument{}
	d.documentReference, err = gojsonreference.NewJsonReference(documentReferenceString)
	d.pool = NewSchemaPool()

	spd, err := d.pool.GetPoolDocument(d.documentReference)
	if err != nil {
		return nil, err
	}

	err = d.parse(spd.Document)
	return &d, err
}

type JsonSchemaDocument struct {
	documentReference gojsonreference.JsonReference
	rootSchema        *JsonSchema
	pool              *SchemaPool
}

func (d *JsonSchemaDocument) parse(document interface{}) error {
	d.rootSchema = &JsonSchema{property: ROOT_SCHEMA_PROPERTY}
	return d.parseSchema(document, d.rootSchema)
}

func (d *JsonSchemaDocument) parseSchema(documentNode interface{}, currentSchema *JsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_OBJECT, STRING_SCHEMA))
	}

	m := documentNode.(map[string]interface{})

	if currentSchema == d.rootSchema {
		if !existsMapKey(m, KEY_SCHEMA) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_IS_REQUIRED, KEY_SCHEMA))
		}
		if !isKind(m[KEY_SCHEMA], reflect.String) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_SCHEMA, STRING_STRING))
		}
		schemaRef := m[KEY_SCHEMA].(string)
		schemaReference, err := gojsonreference.NewJsonReference(schemaRef)
		currentSchema.schema = &schemaReference
		if err != nil {
			return err
		}

		currentSchema.ref = &d.documentReference

		if existsMapKey(m, KEY_REF) {
			return errors.New(fmt.Sprintf("No %s is allowed in root schema", KEY_REF))
		}

	}

	// ref
	if existsMapKey(m, KEY_REF) && !isKind(m[KEY_REF], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_REF, STRING_STRING))
	}
	if k, ok := m[KEY_REF].(string); ok {
		jsonReference, err := gojsonreference.NewJsonReference(k)
		if err != nil {
			return err
		}

		if jsonReference.HasFullUrl {
			currentSchema.ref = &jsonReference
		} else {
			inheritedReference, err := gojsonreference.Inherits(*currentSchema.ref, jsonReference)
			if err != nil {
				return err
			}
			currentSchema.ref = inheritedReference
		}

		dsp, err := d.pool.GetPoolDocument(*currentSchema.ref)
		if err != nil {
			return err
		}

		jsonPointer := currentSchema.ref.GetPointer()

		httpDocumentNode, err := jsonPointer.Get(dsp.Document)
		if err != nil {
			return err
		}

		if !isKind(httpDocumentNode, reflect.Map) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_OBJECT, STRING_SCHEMA))
		}
		m = httpDocumentNode.(map[string]interface{})
	}

	// id
	if existsMapKey(m, KEY_ID) && !isKind(m[KEY_ID], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ID, STRING_STRING))
	}
	if k, ok := m[KEY_ID].(string); ok {
		currentSchema.id = &k
	}

	// title
	if existsMapKey(m, KEY_TITLE) && !isKind(m[KEY_TITLE], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TITLE, STRING_STRING))
	}
	if k, ok := m[KEY_TITLE].(string); ok {
		currentSchema.title = &k
	}

	// description
	if existsMapKey(m, KEY_DESCRIPTION) && !isKind(m[KEY_DESCRIPTION], reflect.String) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_DESCRIPTION, STRING_STRING))
	}
	if k, ok := m[KEY_DESCRIPTION].(string); ok {
		currentSchema.description = &k
	}

	// type
	if !existsMapKey(m, KEY_TYPE) {
		return errors.New(fmt.Sprintf("schema %s - %s is required", currentSchema.property, KEY_TYPE))
	}

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
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, STRING_STRING+"/"+STRING_ARRAY_OF_STRINGS))
				} else {
					currentSchema.types.Add(typeInArray.(string))
				}
			}

		} else {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_TYPE, STRING_STRING+"/"+STRING_ARRAY_OF_STRINGS))
		}
	}

	// properties
	if currentSchema.types.HasType(TYPE_OBJECT) {
		if !existsMapKey(m, KEY_PROPERTIES) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_IS_REQUIRED, KEY_PROPERTIES))
		}

		for k := range m {
			if k == KEY_PROPERTIES {
				err := d.parseProperties(m[k], currentSchema)
				if err != nil {
					return err
				}
			}
		}
	}

	// items
	if currentSchema.types.HasType(TYPE_ARRAY) {
		if !existsMapKey(m, KEY_ITEMS) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_IS_REQUIRED, KEY_ITEMS))
		}

		for k := range m {
			if k == KEY_ITEMS {
				newSchema := &JsonSchema{parent: currentSchema}
				currentSchema.AddPropertiesChild(newSchema)
				err := d.parseSchema(m[k], newSchema)
				if err != nil {
					return err
				}
			}
		}
	}

	// validation : number / integer

	if existsMapKey(m, "multipleOf") {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m["multipleOf"], reflect.Float64) {
				multipleOfValue := m["multipleOf"].(float64)
				if multipleOfValue <= 0 {
					return errors.New("multipleOf must be strictly greater than 0")
				}
				currentSchema.multipleOf = &multipleOfValue
			} else {
				return errors.New("multipleOf must be a number")
			}
		} else {
			return errors.New("multipleOf applies to number,integer")
		}
	}

	if existsMapKey(m, "minimum") {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m["minimum"], reflect.Float64) {
				minimumValue := m["minimum"].(float64)
				currentSchema.minimum = &minimumValue
			} else {
				return errors.New("minimum must be a number")
			}
		} else {
			return errors.New("minimum applies to number,integer")
		}
	}

	if existsMapKey(m, "exclusiveMinimum") {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m["exclusiveMinimum"], reflect.Bool) {
				if currentSchema.minimum == nil {
					return errors.New("exclusiveMinimum cannot exist without maximum")
				}
				exclusiveMinimumValue := m["exclusiveMinimum"].(bool)
				currentSchema.exclusiveMinimum = exclusiveMinimumValue
			} else {
				return errors.New("exclusiveMinimum must be a boolean")
			}
		} else {
			return errors.New("exclusiveMinimum applies to number,integer")
		}
	}

	if currentSchema.minimum != nil && currentSchema.maximum != nil {
		if *currentSchema.minimum > *currentSchema.maximum {
			return errors.New("minimum cannot be greater than maximum")
		}
	}

	// validation : string

	if existsMapKey(m, "minLength") {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m["minLength"], reflect.Float64) {
				minLengthValue := m["minLength"].(float64)
				if isFloat64AnInteger(minLengthValue) {
					if minLengthValue < 0 {
						return errors.New("minLength must be greater than or equal to 0")
					}
					minLengthIntegerValue := int(minLengthValue)
					currentSchema.minLength = &minLengthIntegerValue
				} else {
					return errors.New("minLength must be an integer")
				}
			} else {
				return errors.New("minLength must be an integer")
			}
		} else {
			return errors.New("minLength applies to string")
		}
	}

	if existsMapKey(m, "maxLength") {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m["maxLength"], reflect.Float64) {
				maxLengthValue := m["maxLength"].(float64)
				if isFloat64AnInteger(maxLengthValue) {
					if maxLengthValue < 0 {
						return errors.New("maxLength must be greater than or equal to 0")
					}
					maxLengthIntegerValue := int(maxLengthValue)
					currentSchema.maxLength = &maxLengthIntegerValue
				} else {
					return errors.New("maxLength must be an integer")
				}
			} else {
				return errors.New("maxLength must be an integer")
			}
		} else {
			return errors.New("maxLength applies to string")
		}
	}

	if currentSchema.minLength != nil && currentSchema.maxLength != nil {
		if *currentSchema.minLength > *currentSchema.maxLength {
			return errors.New("minLength cannot be greater than maxLength")
		}
	}

	if existsMapKey(m, "pattern") {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m["pattern"], reflect.String) {
				regexpObject, err := regexp.Compile(m["pattern"].(string))
				if err != nil {
					return errors.New("pattern must be a valid regular expression")
				}
				currentSchema.pattern = regexpObject
			} else {
				return errors.New("pattern must be a string")
			}
		} else {
			return errors.New("pattern applies to string")
		}
	}

	// validation : object

	if existsMapKey(m, "minProperties") {
		if currentSchema.types.HasType(TYPE_OBJECT) {
			if isKind(m["minProperties"], reflect.Float64) {
				minPropertiesValue := m["minProperties"].(float64)
				if isFloat64AnInteger(minPropertiesValue) {
					if minPropertiesValue < 0 {
						return errors.New("minProperties must be greater than or equal to 0")
					}
					minPropertiesntegerValue := int(minPropertiesValue)
					currentSchema.minProperties = &minPropertiesntegerValue
				} else {
					return errors.New("minProperties must be an integer")
				}
			} else {
				return errors.New("minProperties must be an integer")
			}
		} else {
			return errors.New("minProperties applies to string")
		}
	}

	if existsMapKey(m, "maxProperties") {
		if currentSchema.types.HasType(TYPE_OBJECT) {
			if isKind(m["maxProperties"], reflect.Float64) {
				maxPropertiesValue := m["maxProperties"].(float64)
				if isFloat64AnInteger(maxPropertiesValue) {
					if maxPropertiesValue < 0 {
						return errors.New("maxProperties must be greater than or equal to 0")
					}
					maxPropertiesntegerValue := int(maxPropertiesValue)
					currentSchema.maxProperties = &maxPropertiesntegerValue
				} else {
					return errors.New("maxProperties must be an integer")
				}
			} else {
				return errors.New("maxProperties must be an integer")
			}
		} else {
			return errors.New("maxProperties applies to string")
		}
	}

	if currentSchema.minProperties != nil && currentSchema.maxProperties != nil {
		if *currentSchema.minProperties > *currentSchema.maxProperties {
			return errors.New("minProperties cannot be greater than maxProperties")
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseProperties(documentNode interface{}, currentSchema *JsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_AN_OBJECT, STRING_PROPERTIES))
	}

	m := documentNode.(map[string]interface{})
	for k := range m {
		schemaProperty := k
		newSchema := &JsonSchema{property: schemaProperty, parent: currentSchema, ref: currentSchema.ref}
		currentSchema.AddPropertiesChild(newSchema)
		err := d.parseSchema(m[k], newSchema)
		if err != nil {
			return err
		}
	}

	return nil
}
