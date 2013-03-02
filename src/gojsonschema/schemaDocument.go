// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Defines schemaDocument, the main entry to every schemas.
//					Contains the parsing logic and error checking.			
// 
// created      	26-02-2013

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
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, STRING_OBJECT))
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

		httpDocumentNode, _, err := jsonPointer.Get(dsp.Document)
		if err != nil {
			return err
		}

		if !isKind(httpDocumentNode, reflect.Map) {
			return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_SCHEMA, STRING_OBJECT))
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
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_IS_REQUIRED, KEY_TYPE))
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
				if !isKind(m[k], reflect.Map) {
					return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, KEY_ITEMS, STRING_OBJECT))
				}
				newSchema := &JsonSchema{parent: currentSchema, property: k}
				currentSchema.SetItemsChild(newSchema)
				err := d.parseSchema(m[k], newSchema)
				if err != nil {
					return err
				}
			}
		}
	}

	// validation : number / integer

	if existsMapKey(m, KEY_MULTIPLE_OF) {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m[KEY_MULTIPLE_OF], reflect.Float64) {
				multipleOfValue := m[KEY_MULTIPLE_OF].(float64)
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

	if existsMapKey(m, KEY_MINIMUM) {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m[KEY_MINIMUM], reflect.Float64) {
				minimumValue := m[KEY_MINIMUM].(float64)
				currentSchema.minimum = &minimumValue
			} else {
				return errors.New("minimum must be a number")
			}
		} else {
			return errors.New("minimum applies to number,integer")
		}
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MINIMUM) {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m[KEY_EXCLUSIVE_MINIMUM], reflect.Bool) {
				if currentSchema.minimum == nil {
					return errors.New("exclusiveMinimum cannot exist without maximum")
				}
				exclusiveMinimumValue := m[KEY_EXCLUSIVE_MINIMUM].(bool)
				currentSchema.exclusiveMinimum = exclusiveMinimumValue
			} else {
				return errors.New("exclusiveMinimum must be a boolean")
			}
		} else {
			return errors.New("exclusiveMinimum applies to number,integer")
		}
	}

	if existsMapKey(m, KEY_MAXIMUM) {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m[KEY_MAXIMUM], reflect.Float64) {
				maximumValue := m[KEY_MAXIMUM].(float64)
				currentSchema.maximum = &maximumValue
			} else {
				return errors.New("maximum must be a number")
			}
		} else {
			return errors.New("maximum applies to number,integer")
		}
	}

	if existsMapKey(m, KEY_EXCLUSIVE_MAXIMUM) {
		if currentSchema.types.HasType(TYPE_NUMBER) || currentSchema.types.HasType(TYPE_INTEGER) {
			if isKind(m[KEY_EXCLUSIVE_MAXIMUM], reflect.Bool) {
				if currentSchema.maximum == nil {
					return errors.New("exclusiveMaximum cannot exist without maximum")
				}
				exclusiveMaximumValue := m[KEY_EXCLUSIVE_MAXIMUM].(bool)
				currentSchema.exclusiveMaximum = exclusiveMaximumValue
			} else {
				return errors.New("exclusiveMaximum must be a boolean")
			}
		} else {
			return errors.New("exclusiveMaximum applies to number,integer")
		}
	}

	if currentSchema.minimum != nil && currentSchema.maximum != nil {
		if *currentSchema.minimum > *currentSchema.maximum {
			return errors.New("minimum cannot be greater than maximum")
		}
	}

	// validation : string

	if existsMapKey(m, KEY_MIN_LENGTH) {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m[KEY_MIN_LENGTH], reflect.Float64) {
				minLengthValue := m[KEY_MIN_LENGTH].(float64)
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

	if existsMapKey(m, KEY_MAX_LENGTH) {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m[KEY_MAX_LENGTH], reflect.Float64) {
				maxLengthValue := m[KEY_MAX_LENGTH].(float64)
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

	if existsMapKey(m, KEY_PATTERN) {
		if currentSchema.types.HasType(TYPE_STRING) {
			if isKind(m[KEY_PATTERN], reflect.String) {
				regexpObject, err := regexp.Compile(m[KEY_PATTERN].(string))
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

	if existsMapKey(m, KEY_MIN_PROPERTIES) {
		if currentSchema.types.HasType(TYPE_OBJECT) {
			if isKind(m[KEY_MIN_PROPERTIES], reflect.Float64) {
				minPropertiesValue := m[KEY_MIN_PROPERTIES].(float64)
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
			return errors.New("minProperties applies to object")
		}
	}

	if existsMapKey(m, KEY_MAX_PROPERTIES) {
		if currentSchema.types.HasType(TYPE_OBJECT) {
			if isKind(m[KEY_MAX_PROPERTIES], reflect.Float64) {
				maxPropertiesValue := m[KEY_MAX_PROPERTIES].(float64)
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
			return errors.New("maxProperties applies to object")
		}
	}

	if currentSchema.minProperties != nil && currentSchema.maxProperties != nil {
		if *currentSchema.minProperties > *currentSchema.maxProperties {
			return errors.New("minProperties cannot be greater than maxProperties")
		}
	}

	if existsMapKey(m, KEY_REQUIRED) {
		if currentSchema.types.HasType(TYPE_OBJECT) {
			if isKind(m[KEY_REQUIRED], reflect.Slice) {
				requiredValues := m[KEY_REQUIRED].([]interface{})
				for _, requiredValue := range requiredValues {
					if isKind(requiredValue, reflect.String) {
						err := currentSchema.AddRequired(requiredValue.(string))
						if err != nil {
							return err
						}
					} else {
						return errors.New("required items must be string")
					}
				}
			} else {
				return errors.New("required must be an array")
			}
		} else {
			return errors.New("required applies to object")
		}
	}

	// validation : array

	if existsMapKey(m, KEY_MIN_ITEMS) {
		if currentSchema.types.HasType(TYPE_ARRAY) {
			if isKind(m[KEY_MIN_ITEMS], reflect.Float64) {
				minItemsValue := m[KEY_MIN_ITEMS].(float64)
				if isFloat64AnInteger(minItemsValue) {
					if minItemsValue < 0 {
						return errors.New("minItems must be greater than or equal to 0")
					}
					minItemsIntegerValue := int(minItemsValue)
					currentSchema.minItems = &minItemsIntegerValue
				} else {
					return errors.New("minItems must be an integer")
				}
			} else {
				return errors.New("minItems must be an integer")
			}
		} else {
			return errors.New("minItems applies to array")
		}
	}

	if existsMapKey(m, KEY_MAX_ITEMS) {
		if currentSchema.types.HasType(TYPE_ARRAY) {
			if isKind(m[KEY_MAX_ITEMS], reflect.Float64) {
				maxItemsValue := m[KEY_MAX_ITEMS].(float64)
				if isFloat64AnInteger(maxItemsValue) {
					if maxItemsValue < 0 {
						return errors.New("maxItems must be greater than or equal to 0")
					}
					maxItemsIntegerValue := int(maxItemsValue)
					currentSchema.maxItems = &maxItemsIntegerValue
				} else {
					return errors.New("maxItems must be an integer")
				}
			} else {
				return errors.New("maxItems must be an integer")
			}
		} else {
			return errors.New("maxItems applies to array")
		}
	}

	if existsMapKey(m, KEY_UNIQUE_ITEMS) {
		if currentSchema.types.HasType(TYPE_ARRAY) {
			if isKind(m[KEY_UNIQUE_ITEMS], reflect.Bool) {
				currentSchema.uniqueItems = m[KEY_UNIQUE_ITEMS].(bool)
			} else {
				return errors.New("uniqueItems must be an boolean")
			}
		} else {
			return errors.New("uniqueItems applies to array")
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
			return errors.New("enum must be an array")
		}
	}

	// validation : schema

	if existsMapKey(m, KEY_ONE_OF) {
		if isKind(m[KEY_ONE_OF], reflect.Slice) {
			for _, v := range m[KEY_ONE_OF].([]interface{}) {
				newSchema := &JsonSchema{property: KEY_ONE_OF, parent: currentSchema, ref: currentSchema.ref}
				currentSchema.AddOneOf(newSchema)
				err := d.parseSchema(v, newSchema)
				if err != nil {
					return err
				}
			}
		} else {
			return errors.New("oneOf must be an array")
		}
	}

	return nil
}

func (d *JsonSchemaDocument) parseProperties(documentNode interface{}, currentSchema *JsonSchema) error {

	if !isKind(documentNode, reflect.Map) {
		return errors.New(fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, STRING_PROPERTIES, STRING_OBJECT))
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
