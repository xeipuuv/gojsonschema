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
	"regexp"
	"strconv"
	"strings"
)

type ValidationResult struct {
	errorMessages []string
}

func (v *ValidationResult) IsValid() bool {
	return len(v.errorMessages) == 0
}

func (v *ValidationResult) GetErrorMessages() []string {
	return v.errorMessages
}

// Used to copy errors from a sub-schema validation to the main one
func (v *ValidationResult) CopyErrorMessages(others []string) {
	v.errorMessages = append(v.errorMessages, others...)
}

func (v *ValidationResult) CopyErrorMessagesWithAnnotation(annotation string, others []string) {
	for i := range others {
		v.errorMessages = append(v.errorMessages, annotation+` `+others[i])
	}
}

func (v *ValidationResult) addErrorMessage(context *jsonContext, message string) {
	fullMessage := fmt.Sprintf("%v : %v", context, message)
	v.errorMessages = append(v.errorMessages, fullMessage)
}

func (v *JsonSchemaDocument) Validate(document interface{}) ValidationResult {

	result := ValidationResult{}
	context := consJsonContext("ROOT", nil)
	v.rootSchema.validateRecursive(v.rootSchema, document, &result, context)
	return result
}

func (v *jsonSchema) Validate(document interface{}, context *jsonContext) ValidationResult {

	result := ValidationResult{}
	v.validateRecursive(v, document, &result, context)
	return result
}

// Walker function to validate the json recursively against the schema
func (v *jsonSchema) validateRecursive(currentSchema *jsonSchema, currentNode interface{}, result *ValidationResult, context *jsonContext) {

	// Handle referenced schemas, returns directly when a $ref is found
	if currentSchema.refSchema != nil {
		v.validateRecursive(currentSchema.refSchema, currentNode, result, context)
		return
	}

	// Check for null value
	if currentNode == nil {
		if currentSchema.types.HasTypeInSchema() && !currentSchema.types.HasType(TYPE_NULL) {
			result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
			return
		}

		currentSchema.validateSchema(currentSchema, currentNode, result, context)
		v.validateCommon(currentSchema, currentNode, result, context)
	} else { // Not null value :

		rValue := reflect.ValueOf(currentNode)
		rKind := rValue.Kind()

		switch rKind {

		// Slice => JSON array

		case reflect.Slice:

			if currentSchema.types.HasTypeInSchema() && !currentSchema.types.HasType(TYPE_ARRAY) {
				result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
				return
			}

			castCurrentNode := currentNode.([]interface{})

			currentSchema.validateSchema(currentSchema, castCurrentNode, result, context)

			v.validateArray(currentSchema, castCurrentNode, result, context)
			v.validateCommon(currentSchema, castCurrentNode, result, context)

		// Map => JSON object

		case reflect.Map:
			if currentSchema.types.HasTypeInSchema() && !currentSchema.types.HasType(TYPE_OBJECT) {
				result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
				return
			}

			castCurrentNode := currentNode.(map[string]interface{})

			currentSchema.validateSchema(currentSchema, castCurrentNode, result, context)

			v.validateObject(currentSchema, castCurrentNode, result, context)
			v.validateCommon(currentSchema, castCurrentNode, result, context)

			for _, pSchema := range currentSchema.propertiesChildren {
				nextNode, ok := castCurrentNode[pSchema.property]
				if ok {
					subContext := consJsonContext(pSchema.property, context)
					v.validateRecursive(pSchema, nextNode, result, subContext)
				}
			}

		// Simple JSON values : string, number, boolean

		case reflect.Bool:

			if currentSchema.types.HasTypeInSchema() && !currentSchema.types.HasType(TYPE_BOOLEAN) {
				result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
				return
			}

			value := currentNode.(bool)

			currentSchema.validateSchema(currentSchema, value, result, context)
			v.validateNumber(currentSchema, value, result, context)
			v.validateCommon(currentSchema, value, result, context)
			v.validateString(currentSchema, value, result, context)

		case reflect.String:

			if currentSchema.types.HasTypeInSchema() && !currentSchema.types.HasType(TYPE_STRING) {
				result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
				return
			}

			value := currentNode.(string)

			currentSchema.validateSchema(currentSchema, value, result, context)
			v.validateNumber(currentSchema, value, result, context)
			v.validateCommon(currentSchema, value, result, context)
			v.validateString(currentSchema, value, result, context)

		case reflect.Float64:

			value := currentNode.(float64)

			// Note: JSON only understand one kind of numeric ( can be float or int )
			// JSON schema make a distinction between fload and int
			// An integer can be a number, but a number ( with decimals ) cannot be an integer
			// Here is the test:
			isInteger := isFloat64AnInteger(value) // "weird" (?) thing: Go's Atoi accepts 1.0, 45.0 as integers...

			formatIsCorrect := currentSchema.types.HasType(TYPE_NUMBER) || (isInteger && currentSchema.types.HasType(TYPE_INTEGER))

			if currentSchema.types.HasTypeInSchema() && !formatIsCorrect {
				result.addErrorMessage(context, fmt.Sprintf(ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y, currentSchema.property, currentSchema.types.String()))
				return
			}

			currentSchema.validateSchema(currentSchema, value, result, context)
			v.validateNumber(currentSchema, value, result, context)
			v.validateCommon(currentSchema, value, result, context)
			v.validateString(currentSchema, value, result, context)
		}
	}
}

// Different kinds of validation there, schema / common / array / object / string...
// Again, this is pretty straight forward and simple to understand

func (v *jsonSchema) validateSchema(currentSchema *jsonSchema, currentNode interface{}, result *ValidationResult, context *jsonContext) {

	if len(currentSchema.anyOf) > 0 {
		validatedAnyOf := false

		for _, anyOfSchema := range currentSchema.anyOf {
			if !validatedAnyOf {
				validationResult := anyOfSchema.Validate(currentNode, context)
				validatedAnyOf = validationResult.IsValid()
			}
		}
		if !validatedAnyOf {
			result.addErrorMessage(context, fmt.Sprintf("%s failed to validate any of the schema", currentSchema.property))
		}
	}

	if len(currentSchema.oneOf) > 0 {
		nbValidated := 0

		for _, oneOfSchema := range currentSchema.oneOf {
			validationResult := oneOfSchema.Validate(currentNode, context)
			if validationResult.IsValid() {
				nbValidated++
			}
		}

		if nbValidated != 1 {
			result.addErrorMessage(context, fmt.Sprintf("%s failed to validate one of the schema", currentSchema.property))
		}
	}

	if len(currentSchema.allOf) > 0 {
		nbValidated := 0

		for _, allOfSchema := range currentSchema.allOf {
			validationResult := allOfSchema.Validate(currentNode, context)
			if validationResult.IsValid() {
				nbValidated++
			} else {
				result.CopyErrorMessages(validationResult.GetErrorMessages())
			}
		}

		if nbValidated != len(currentSchema.allOf) {
			result.addErrorMessage(context, fmt.Sprintf("%s failed to validate all of the schema", currentSchema.property))
		}
	}

	if currentSchema.not != nil {
		validationResult := currentSchema.not.Validate(currentNode, context)
		if validationResult.IsValid() {
			result.addErrorMessage(context, fmt.Sprintf("%s is not allowed to validate the schema", currentSchema.property))
		}
	}

	if currentSchema.dependencies != nil && len(currentSchema.dependencies) > 0 {
		if isKind(currentNode, reflect.Map) {
			for elementKey := range currentNode.(map[string]interface{}) {
				if dependency, ok := currentSchema.dependencies[elementKey]; ok {
					switch dependency := dependency.(type) {

					case []string:
						for _, dependOnKey := range dependency {
							if _, dependencyResolved := currentNode.(map[string]interface{})[dependOnKey]; !dependencyResolved {
								result.addErrorMessage(context, fmt.Sprintf("%s has a dependency on %s", elementKey, dependOnKey))
							}
						}

					case *jsonSchema:
						dependency.validateRecursive(dependency, currentNode, result, context)

					}
				}
			}
		}
	}
}

func (v *jsonSchema) validateCommon(currentSchema *jsonSchema, value interface{}, result *ValidationResult, context *jsonContext) {

	if len(currentSchema.enum) > 0 {
		has, err := currentSchema.HasEnum(value)
		if err != nil {
			result.addErrorMessage(context, err.Error())
		}
		if !has {
			result.addErrorMessage(context, fmt.Sprintf("%s must match one of the enum values [%s]", currentSchema.property, strings.Join(currentSchema.enum, ",")))
		}
	}
}

func (v *jsonSchema) validateArray(currentSchema *jsonSchema, value []interface{}, result *ValidationResult, context *jsonContext) {

	nbItems := len(value)

	if currentSchema.itemsChildrenIsSingleSchema {
		for i := range value {
			subContext := consJsonContext(strconv.Itoa(i), context)
			validationResult := currentSchema.itemsChildren[0].Validate(value[i], subContext)
			if !validationResult.IsValid() {
				result.CopyErrorMessagesWithAnnotation(currentSchema.property, validationResult.GetErrorMessages())
			}
		}
	} else {
		if currentSchema.itemsChildren != nil && len(currentSchema.itemsChildren) > 0 {

			nbItems := len(currentSchema.itemsChildren)
			nbValues := len(value)

			if nbItems == nbValues {
				for i := 0; i != nbItems; i++ {
					subContext := consJsonContext(strconv.Itoa(i), context)
					validationResult := currentSchema.itemsChildren[i].Validate(value[i], subContext)
					if !validationResult.IsValid() {
						result.CopyErrorMessages(validationResult.GetErrorMessages())
					}
				}
			} else if nbItems < nbValues {
				switch currentSchema.additionalItems.(type) {
				case bool:
					if !currentSchema.additionalItems.(bool) {
						result.addErrorMessage(context, fmt.Sprintf("No additional item allowed on %s", currentSchema.property))
					}
				case *jsonSchema:
					additionalItemSchema := currentSchema.additionalItems.(*jsonSchema)
					for i := nbItems; i != nbValues; i++ {
						subContext := consJsonContext(strconv.Itoa(i), context)
						validationResult := additionalItemSchema.Validate(value[i], subContext)
						if !validationResult.IsValid() {
							result.CopyErrorMessages(validationResult.GetErrorMessages())
						}
					}

				}
			}
		}
	}

	if currentSchema.minItems != nil {
		if nbItems < *currentSchema.minItems {
			result.addErrorMessage(context, fmt.Sprintf("%s must have at least %d items", currentSchema.property, *currentSchema.minItems))
		}
	}

	if currentSchema.maxItems != nil {
		if nbItems > *currentSchema.maxItems {
			result.addErrorMessage(context, fmt.Sprintf("%s must have at the most %d items", currentSchema.property, *currentSchema.maxItems))
		}
	}

	if currentSchema.uniqueItems {
		var stringifiedItems []string
		for _, v := range value {
			vString, err := marshalToString(v)
			if err != nil {
				result.addErrorMessage(context, fmt.Sprintf("%s could not be marshalled", currentSchema.property))
			}
			if isStringInSlice(stringifiedItems, *vString) {
				result.addErrorMessage(context, fmt.Sprintf("%s items must be unique", currentSchema.property))
			}
			stringifiedItems = append(stringifiedItems, *vString)
		}
	}

}

func (v *jsonSchema) validateObject(currentSchema *jsonSchema, value map[string]interface{}, result *ValidationResult, context *jsonContext) {

	if currentSchema.minProperties != nil {
		if len(value) < *currentSchema.minProperties {
			result.addErrorMessage(context, fmt.Sprintf("%s must have at least %d properties", currentSchema.property, *currentSchema.minProperties))
		}
	}

	if currentSchema.maxProperties != nil {
		if len(value) > *currentSchema.maxProperties {
			result.addErrorMessage(context, fmt.Sprintf("%s must have at the most %d properties", currentSchema.property, *currentSchema.maxProperties))
		}
	}

	for _, requiredProperty := range currentSchema.required {
		_, ok := value[requiredProperty]
		if !ok {
			result.addErrorMessage(context, fmt.Sprintf("%s property is required", requiredProperty))
		}
	}

	if currentSchema.additionalProperties != nil {
		switch currentSchema.additionalProperties.(type) {
		case bool:
			if !currentSchema.additionalProperties.(bool) {
				for pk := range value {
					found := false
					for _, spValue := range currentSchema.propertiesChildren {
						if pk == spValue.property {
							found = true
						}
					}

					if !found {
						result.addErrorMessage(context, fmt.Sprintf("No additional property ( %s ) is allowed on %s", pk, currentSchema.property))
					}
				}
			}

		case *jsonSchema:
			additionalPropertiesSchema := currentSchema.additionalProperties.(*jsonSchema)
			for pk := range value {
				found := false
				for _, spValue := range currentSchema.propertiesChildren {
					if pk == spValue.property {
						found = true
					}
				}
				if !found {

					overridePatternPropertiesMatches := false

					// check patternProperties on not found one since patternProperties overrides
					if currentSchema.patternProperties != nil {
						for ovk := range value {
							for ppk, ppv := range currentSchema.patternProperties {
								if matches, _ := regexp.MatchString(ppk, ovk); matches {
									subContext := consJsonContext(ovk, context)
									validationResult := ppv.Validate(value[ovk], subContext)
									if !validationResult.IsValid() {
										result.CopyErrorMessages(validationResult.GetErrorMessages())
									} else {
										overridePatternPropertiesMatches = true
									}
								}
							}
						}
					}

					// both additionalProperties and patternProperties failed
					if !overridePatternPropertiesMatches {
						validationResult := additionalPropertiesSchema.Validate(value[pk], context)
						if !validationResult.IsValid() {
							result.CopyErrorMessages(validationResult.GetErrorMessages())
						}
					}
				}
			}

		}
	}

	if currentSchema.patternProperties != nil {
		for k := range value {
			for pk, pv := range currentSchema.patternProperties {
				if matches, _ := regexp.MatchString(pk, k); matches {
					subContext := consJsonContext(k, context)
					validationResult := pv.Validate(value[k], subContext)
					if !validationResult.IsValid() {
						result.CopyErrorMessages(validationResult.GetErrorMessages())
					}
				}
			}
		}
	}

}

func (v *jsonSchema) validateString(currentSchema *jsonSchema, value interface{}, result *ValidationResult, context *jsonContext) {

	// Ignore non strings
	if !isKind(value, reflect.String) {
		return
	}

	stringValue := value.(string)

	if currentSchema.minLength != nil {
		if len(stringValue) < *currentSchema.minLength {
			result.addErrorMessage(context, fmt.Sprintf("%s's length must be greater or equal to %d", currentSchema.property, *currentSchema.minLength))
		}
	}

	if currentSchema.maxLength != nil {
		if len(stringValue) > *currentSchema.maxLength {
			result.addErrorMessage(context, fmt.Sprintf("%s's length must be lower or equal to %d", currentSchema.property, *currentSchema.maxLength))
		}
	}

	if currentSchema.pattern != nil {
		if !currentSchema.pattern.MatchString(stringValue) {
			result.addErrorMessage(context, fmt.Sprintf("%s has an invalid format", currentSchema.property))
		}
	}

}

func (v *jsonSchema) validateNumber(currentSchema *jsonSchema, value interface{}, result *ValidationResult, context *jsonContext) {

	// Ignore non numbers
	if !isKind(value, reflect.Float64) {
		return
	}

	float64Value := value.(float64)

	if currentSchema.multipleOf != nil {
		if !isFloat64AnInteger(float64Value / *currentSchema.multipleOf) {
			result.addErrorMessage(context, fmt.Sprintf("%s (%s) is not a multiple of %s", currentSchema.property, validationErrorFormatNumber(float64Value), validationErrorFormatNumber(*currentSchema.multipleOf)))
		}
	}

	if currentSchema.maximum != nil {
		if currentSchema.exclusiveMaximum {
			if float64Value >= *currentSchema.maximum {
				result.addErrorMessage(context, fmt.Sprintf("%s (%s) must be lower than or equal to %s", currentSchema.property, validationErrorFormatNumber(float64Value), validationErrorFormatNumber(*currentSchema.maximum)))
			}
		} else {
			if float64Value > *currentSchema.maximum {
				result.addErrorMessage(context, fmt.Sprintf("%s (%s) must be lower than %s", currentSchema.property, validationErrorFormatNumber(float64Value), validationErrorFormatNumber(*currentSchema.maximum)))
			}
		}
	}

	if currentSchema.minimum != nil {
		if currentSchema.exclusiveMinimum {
			if float64Value <= *currentSchema.minimum {
				result.addErrorMessage(context, fmt.Sprintf("%s (%s) must be greater than or equal to %s", currentSchema.property, validationErrorFormatNumber(float64Value), validationErrorFormatNumber(*currentSchema.minimum)))
			}
		} else {
			if float64Value < *currentSchema.minimum {
				result.addErrorMessage(context, fmt.Sprintf("%s (%s) must be greater than %s", currentSchema.property, validationErrorFormatNumber(float64Value), validationErrorFormatNumber(*currentSchema.minimum)))
			}
		}
	}

}
