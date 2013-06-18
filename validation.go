// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
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

// author           sigu-399
// author-github    https://github.com/sigu-399
// author-mail      sigu.399@gmail.com
//
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
//
// description      Extends JsonSchemaDocument and jsonSchema, implements the validation phase.
//
// created          28-02-2013

package gojsonschema

import (
	"fmt"
	"reflect"
)

type ValidationResult struct {
	valid         bool
	errorMessages []string
}

func (v *ValidationResult) IsValid() bool {
	return v.valid
}

func (v *ValidationResult) GetErrorMessages() []string {
	return v.errorMessages
}

func (v *ValidationResult) CopyErrorMessages(others []string) {
	v.errorMessages = append(v.errorMessages, others...)
}

func (v *ValidationResult) addErrorMessage(message string) {
	v.errorMessages = append(v.errorMessages, message)
	v.valid = false
}

func (v *JsonSchemaDocument) Validate(document interface{}) ValidationResult {

	result := ValidationResult{valid: true}
	v.rootSchema.validateRecursive(v.rootSchema, document, &result)
	return result
}

func (v *jsonSchema) Validate(document interface{}) ValidationResult {

	result := ValidationResult{valid: true}
	v.validateRecursive(v, document, &result)
	return result
}

// Walker function to validate the json recursively against the schema
func (v *jsonSchema) validateRecursive(currentSchema *jsonSchema, currentNode interface{}, result *ValidationResult) {

	schProperty := currentSchema.property
	schTypes := currentSchema.types

	if currentNode == nil {
		if !schTypes.HasType(TYPE_NULL) {
			result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
			return
		}
	} else {

		rValue := reflect.ValueOf(currentNode)
		rKind := rValue.Kind()

		var nextNode interface{}
		var ok bool
		switch rKind {

		// Slice => JSON array

		case reflect.Slice:

			if schTypes.HasTypeInSchema() && !schTypes.HasType(TYPE_ARRAY) {
				result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			castCurrentNode := currentNode.([]interface{})

			v.validateArray(currentSchema, castCurrentNode, result)
			v.validateCommon(currentSchema, castCurrentNode, result)

		// Map => JSON object

		case reflect.Map:
			if schTypes.HasTypeInSchema() && !schTypes.HasType(TYPE_OBJECT) {
				result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			castCurrentNode := currentNode.(map[string]interface{})

			currentSchema.validateSchema(currentSchema, castCurrentNode, result)

			v.validateObject(currentSchema, castCurrentNode, result)
			v.validateCommon(currentSchema, castCurrentNode, result)

			for _, pSchema := range currentSchema.propertiesChildren {
				nextNode, ok = castCurrentNode[pSchema.property]
				if ok {
					v.validateRecursive(pSchema, nextNode, result)
				}
			}

		// Simple JSON values : string, number, boolean

		case reflect.Bool:

			if schTypes.HasTypeInSchema() && !schTypes.HasType(TYPE_BOOLEAN) {
				result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			value := currentNode.(bool)

			currentSchema.validateSchema(currentSchema, value, result)
			v.validateNumber(currentSchema, value, result)
			v.validateCommon(currentSchema, value, result)
			v.validateString(currentSchema, value, result)

		case reflect.String:

			if schTypes.HasTypeInSchema() && !schTypes.HasType(TYPE_STRING) {
				result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			value := currentNode.(string)

			currentSchema.validateSchema(currentSchema, value, result)
			v.validateNumber(currentSchema, value, result)
			v.validateCommon(currentSchema, value, result)
			v.validateString(currentSchema, value, result)

		case reflect.Float64:

			value := currentNode.(float64)

			// Note: JSON only understand one kind of numeric ( can be float or int )
			// JSON schema make a distinction between fload and int
			// An integer can be a number, but a number ( with decimals ) cannot be an integer
			// Here is the test:
			isInteger := isFloat64AnInteger(value) // "weird" (?) thing: Go's Atoi accepts 1.0, 45.0 as integers...

			formatIsCorrect := schTypes.HasType(TYPE_NUMBER) || (isInteger && schTypes.HasType(TYPE_INTEGER))

			if schTypes.HasTypeInSchema() && !formatIsCorrect {
				result.addErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			currentSchema.validateSchema(currentSchema, value, result)
			v.validateNumber(currentSchema, value, result)
			v.validateCommon(currentSchema, value, result)
			v.validateString(currentSchema, value, result)
		}
	}
}

// Different kinds of validation there, schema / common / array / object / string...
// Again, this is pretty straight forward and simple to understand

func (v *jsonSchema) validateSchema(currentSchema *jsonSchema, currentNode interface{}, result *ValidationResult) {

	if len(currentSchema.anyOf) > 0 {
		validatedAnyOf := false

		for _, anyOfSchema := range currentSchema.anyOf {
			if !validatedAnyOf {
				validationResult := anyOfSchema.Validate(currentNode)
				validatedAnyOf = validationResult.IsValid()
			}
		}
		if !validatedAnyOf {
			result.addErrorMessage(fmt.Sprintf("%s failed to validate any of the schema", currentSchema.property))
		}
	}

	if len(currentSchema.oneOf) > 0 {
		nbValidated := 0

		for _, oneOfSchema := range currentSchema.oneOf {
			validationResult := oneOfSchema.Validate(currentNode)
			if validationResult.IsValid() {
				nbValidated++
			}
		}

		if nbValidated != 1 {
			result.addErrorMessage(fmt.Sprintf("%s failed to validate one of the schema", currentSchema.property))
		}
	}

	if len(currentSchema.allOf) > 0 {
		nbValidated := 0

		for _, allOfSchema := range currentSchema.allOf {
			validationResult := allOfSchema.Validate(currentNode)
			if validationResult.IsValid() {
				nbValidated++
			} else {
				result.CopyErrorMessages(validationResult.GetErrorMessages())
			}
		}

		if nbValidated != len(currentSchema.allOf) {
			result.addErrorMessage(fmt.Sprintf("%s failed to validate all of the schema", currentSchema.property))
		}
	}

	if currentSchema.not != nil {
		validationResult := currentSchema.not.Validate(currentNode)
		if validationResult.IsValid() {
			result.addErrorMessage(fmt.Sprintf("%s is not allowed to validate the schema", currentSchema.property))
		}
	}

}

func (v *jsonSchema) validateCommon(currentSchema *jsonSchema, value interface{}, result *ValidationResult) {

	if len(currentSchema.enum) > 0 {
		has, err := currentSchema.HasEnum(value)
		if err != nil {
			result.addErrorMessage(err.Error())
		}
		if !has {
			result.addErrorMessage(fmt.Sprintf("%s must match one of the enum values", currentSchema.property))
		}
	}
}

func (v *jsonSchema) validateArray(currentSchema *jsonSchema, value []interface{}, result *ValidationResult) {

	if currentSchema.minItems != nil {
		if len(value) < *currentSchema.minItems {
			result.addErrorMessage(fmt.Sprintf("%s must have at least %d items", currentSchema.property, *currentSchema.minItems))
		}
	}

	if currentSchema.maxItems != nil {
		if len(value) > *currentSchema.maxItems {
			result.addErrorMessage(fmt.Sprintf("%s must have at the most %d items", currentSchema.property, *currentSchema.maxItems))
		}
	}

	if currentSchema.uniqueItems {
		var stringifiedItems []string
		for _, v := range value {
			vString, err := marshalToString(v)
			if err != nil {
				result.addErrorMessage(fmt.Sprintf("%s could not be marshalled", currentSchema.property))
			}
			if isStringInSlice(stringifiedItems, *vString) {
				result.addErrorMessage(fmt.Sprintf("%s items must be unique", currentSchema.property))
			}
			stringifiedItems = append(stringifiedItems, *vString)
		}
	}

}

func (v *jsonSchema) validateObject(currentSchema *jsonSchema, value map[string]interface{}, result *ValidationResult) {

	if currentSchema.minProperties != nil {
		if len(value) < *currentSchema.minProperties {
			result.addErrorMessage(fmt.Sprintf("%s must have at least %d properties", currentSchema.property, *currentSchema.minProperties))
		}
	}

	if currentSchema.maxProperties != nil {
		if len(value) > *currentSchema.maxProperties {
			result.addErrorMessage(fmt.Sprintf("%s must have at the most %d properties", currentSchema.property, *currentSchema.maxProperties))
		}
	}

	for _, requiredProperty := range currentSchema.required {
		_, ok := value[requiredProperty]
		if !ok {
			result.addErrorMessage(fmt.Sprintf("%s property is required", requiredProperty))
		}
	}

}

func (v *jsonSchema) validateString(currentSchema *jsonSchema, value interface{}, result *ValidationResult) {

	reflectValue := reflect.ValueOf(value)
	reflectKind := reflectValue.Kind()

	if currentSchema.minLength != nil {
		if reflectKind == reflect.String {
			if len(value.(string)) < *currentSchema.minLength {
				result.addErrorMessage(fmt.Sprintf("%s's length must be greater or equal to %d", currentSchema.property, *currentSchema.minLength))
			}
		}
	}

	if currentSchema.maxLength != nil {
		if reflectKind == reflect.String {
			if len(value.(string)) > *currentSchema.maxLength {
				result.addErrorMessage(fmt.Sprintf("%s's length must be lower or equal to %d", currentSchema.property, *currentSchema.maxLength))
			}
		}
	}

	if currentSchema.pattern != nil {
		if reflectKind == reflect.String {
			if !currentSchema.pattern.MatchString(value.(string)) {
				result.addErrorMessage(fmt.Sprintf("%s has an invalid format", currentSchema.property))
			}
		}
	}

}

func (v *jsonSchema) validateNumber(currentSchema *jsonSchema, value interface{}, result *ValidationResult) {

	reflectValue := reflect.ValueOf(value)
	reflectKind := reflectValue.Kind()

	if currentSchema.multipleOf != nil {
		if reflectKind == reflect.Float64 {
			if !isFloat64AnInteger(value.(float64) / *currentSchema.multipleOf) {
				result.addErrorMessage(fmt.Sprintf("%f is not a multiple of %f", value.(float64), *currentSchema.multipleOf))
			}
		}
	}

	if currentSchema.maximum != nil {
		if reflectKind == reflect.Float64 {
			if currentSchema.exclusiveMaximum {
				if value.(float64) >= *currentSchema.maximum {
					result.addErrorMessage(fmt.Sprintf("%f must be lower than or equal to %f", value.(float64), *currentSchema.maximum))
				}
			} else {
				if value.(float64) > *currentSchema.maximum {
					result.addErrorMessage(fmt.Sprintf("%f must be lower than %f", value.(float64), *currentSchema.maximum))
				}
			}
		}
	}

	if currentSchema.minimum != nil {
		if reflectKind == reflect.Float64 {
			if currentSchema.exclusiveMinimum {
				if value.(float64) <= *currentSchema.minimum {
					result.addErrorMessage(fmt.Sprintf("%f must be greater than or equal to %f", value.(float64), *currentSchema.minimum))
				}
			} else {
				if value.(float64) < *currentSchema.minimum {
					result.addErrorMessage(fmt.Sprintf("%f must be greater than %f", value.(float64), *currentSchema.minimum))
				}
			}
		}
	}

}
