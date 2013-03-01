// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      28-02-2013

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

func (v *ValidationResult) AddErrorMessage(message string) {
	v.errorMessages = append(v.errorMessages, message)
	v.valid = false
}

func (v *JsonSchemaDocument) Validate(document interface{}) ValidationResult {

	result := ValidationResult{valid: true}
	v.validateRecursive(v.rootSchema, document, &result)
	return result
}

func (v *JsonSchemaDocument) validateRecursive(currentSchema *JsonSchema, currentNode interface{}, result *ValidationResult) {

	fmt.Printf("Validation of schema %s\n", currentSchema.property)

	schProperty := currentSchema.property
	schTypes := currentSchema.types

	if currentNode == nil {
		if !schTypes.HasType(TYPE_NULL) {
			result.AddErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
			return
		}
	} else {

		rValue := reflect.ValueOf(currentNode)
		rKind := rValue.Kind()

		var nextNode interface{}
		var ok bool

		fmt.Printf("Type %s\n", rKind.String())

		switch rKind {

		case reflect.Map:

			if !schTypes.HasType(TYPE_OBJECT) {
				result.AddErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			for _, pSchema := range currentSchema.propertiesChildren {
				castCurrentNode := currentNode.(map[string]interface{})
				nextNode, ok = castCurrentNode[pSchema.property]
				if !ok {
					result.AddErrorMessage(fmt.Sprintf("%s is required", pSchema.property))
					return
				}
				v.validateRecursive(pSchema, nextNode, result)
			}

		case reflect.String:

			if !schTypes.HasType(TYPE_STRING) {
				result.AddErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			value := currentNode.(string)
			v.validateString(currentSchema, value, result)

		case reflect.Float64:

			value := currentNode.(float64)
			isInteger := isFloat64AnInteger(value)

			formatIsCorrect := schTypes.HasType(TYPE_NUMBER) || (isInteger && schTypes.HasType(TYPE_INTEGER))

			if !formatIsCorrect {
				result.AddErrorMessage(fmt.Sprintf("%s must be of type %s", schProperty, schTypes.String()))
				return
			}

			v.validateNumber(currentSchema, value, result)
		}
	}
}

func (v *JsonSchemaDocument) validateString(currentSchema *JsonSchema, value string, result *ValidationResult) {

	if currentSchema.minLength != nil {
		if len(value) < *currentSchema.minLength {
			result.AddErrorMessage(fmt.Sprintf("%s's length must be greater or equal to %d", currentSchema.property, *currentSchema.minLength))
		}
	}
}

func (v *JsonSchemaDocument) validateNumber(currentSchema *JsonSchema, value float64, result *ValidationResult) {

	if currentSchema.multipleOf != nil {
		if !isFloat64AnInteger(value / *currentSchema.multipleOf) {
			result.AddErrorMessage(fmt.Sprintf("%f is not a multiple of %f", value, *currentSchema.multipleOf))
		}
	}

	if currentSchema.maximum != nil {
		if currentSchema.exclusiveMaximum {
			if value > *currentSchema.maximum {
				result.AddErrorMessage(fmt.Sprintf("%f must be lower than or equal to %f", value, *currentSchema.maximum))
			}
		} else {
			if value >= *currentSchema.maximum {
				result.AddErrorMessage(fmt.Sprintf("%f must be lower than %f", value, *currentSchema.maximum))
			}
		}
	}

	if currentSchema.minimum != nil {
		if currentSchema.exclusiveMinimum {
			if value > *currentSchema.minimum {
				result.AddErrorMessage(fmt.Sprintf("%f must be greater than or equal to %f", value, *currentSchema.minimum))
			}
		} else {
			if value >= *currentSchema.minimum {
				result.AddErrorMessage(fmt.Sprintf("%f must be greater than %f", value, *currentSchema.minimum))
			}
		}
	}

}
