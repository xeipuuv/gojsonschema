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
// description      Extends Schema and subSchema, implements the validation phase.
//
// created          28-02-2013

package gojsonschema

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Validate loads and validates a JSON schema
func Validate(ls JSONLoader, ld JSONLoader, evaluator ExpressionEvaluator) (*Result, error) {
	// load schema
	schema, err := NewSchema(ls, evaluator)
	if err != nil {
		return nil, err
	}
	return schema.Validate(ld)
}

// Validate loads and validates a JSON document
func (v *Schema) Validate(l JSONLoader) (*Result, error) {
	root, err := l.LoadJSON()
	if err != nil {
		return nil, err
	}
	return v.validateDocument(root), nil
}

func (v *Schema) validateDocument(root interface{}) *Result {
	result := &Result{}
	context := NewJsonContext(STRING_CONTEXT_ROOT, nil)
	v.rootSchema.validateRecursive(v.rootSchema, root, result, context, v.rootSchema.fieldPath)

	return result
}

func (v *subSchema) subValidateWithContext(document interface{}, context *JsonContext) *Result {
	result := &Result{}
	v.validateRecursive(v, document, result, context, v.fieldPath)
	return result
}

// Walker function to validate the json recursively against the subSchema
func (v *subSchema) validateRecursive(currentSubSchema *subSchema, currentNode interface{}, result *Result, context *JsonContext, fieldPath *[]string) {
	currentSubSchema.fieldPath = v.fieldPath

	if internalLogEnabled {
		internalLog("validateRecursive %s", context.String())
		internalLog(" %v", currentNode)
	}

	// Handle referenced schemas, returns directly when a $ref is found
	if currentSubSchema.refSchema != nil {
		v.validateRecursive(currentSubSchema.refSchema, currentNode, result, context, fieldPath)
		return
	}

	// Check for null value
	if currentNode == nil {
		validType := currentSubSchema.types.Contains(TYPE_NULL) || currentSubSchema.bsonTypes.Contains(TYPE_NULL)
		if (currentSubSchema.types.IsTyped() || currentSubSchema.bsonTypes.IsTyped()) && !validType {
			result.addInternalError(
				new(InvalidTypeError),
				context,
				currentNode,
				ErrorDetails{
					"expected": currentSubSchema.types.String(),
					"given":    TYPE_NULL,
				},
			)
			return
		}

		v.validateSchema(currentSubSchema, currentNode, result, context)
		v.validateCommon(currentSubSchema, currentNode, result, context)

	} else if isJSONNumber(currentNode) {

		value := currentNode.(json.Number)

		isInt := checkJSONInteger(value)

		validType := currentSubSchema.types.Contains(TYPE_NUMBER) || (isInt && currentSubSchema.types.Contains(TYPE_INTEGER))

		if currentSubSchema.types.IsTyped() && !validType {

			givenType := TYPE_INTEGER
			if !isInt {
				givenType = TYPE_NUMBER
			}

			result.addInternalError(
				new(InvalidTypeError),
				context,
				currentNode,
				ErrorDetails{
					"expected": currentSubSchema.types.String(),
					"given":    givenType,
				},
			)
			return
		}

		v.validateSchema(currentSubSchema, value, result, context)
		v.validateNumber(currentSubSchema, value, result, context)
		v.validateCommon(currentSubSchema, value, result, context)
		v.validateString(currentSubSchema, value, result, context)

	} else if isNumber(currentNode) {

		var value interface{}
		switch cn := currentNode.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_INTEGER, TYPE_INT32, TYPE_INT64, TYPE_NUMBER) {
				return
			}
			value = cn
		case float32, float64:
			if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_DOUBLE, TYPE_NUMBER) {
				return
			}
			value = cn
		}

		v.validateSchema(currentSubSchema, value, result, context)
		v.validateNumber(currentSubSchema, value, result, context)
		v.validateCommon(currentSubSchema, value, result, context)
		v.validateString(currentSubSchema, value, result, context)

	} else {

		rValue := reflect.ValueOf(currentNode)
		rKind := rValue.Kind()

		switch rKind {

		// Slice => JSON array
		case reflect.Slice:
			switch cn := currentNode.(type) {

			case bson.D:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_OBJECT) {
					return
				}

				v.validateSchema(currentSubSchema, cn, result, context)
				v.validateObject(currentSubSchema, cn.Map(), result, context)
				v.validateCommon(currentSubSchema, cn, result, context)

				for _, pSchema := range currentSubSchema.propertiesChildren {
					nextNode, ok := cn.Map()[pSchema.property]
					if ok {
						subContext := NewJsonContext(pSchema.property, context)
						*fieldPath = append(*fieldPath, pSchema.property)
						v.validateRecursive(pSchema, nextNode, result, subContext, fieldPath)
						*fieldPath = (*fieldPath)[:len(*fieldPath)-1]
					}
				}

			case []interface{}:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_ARRAY) {
					return
				}
				castCurrentNode := currentNode.([]interface{})

				v.validateSchema(currentSubSchema, castCurrentNode, result, context)
				v.validateArray(currentSubSchema, castCurrentNode, result, context)
				v.validateCommon(currentSubSchema, castCurrentNode, result, context)
			}

		// Map => JSON object
		case reflect.Map:
			if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_OBJECT) {
				return
			}

			castCurrentNode, ok := currentNode.(map[string]interface{})
			if !ok {
				castCurrentNode = convertDocumentNode(currentNode).(map[string]interface{})
			}

			v.validateSchema(currentSubSchema, castCurrentNode, result, context)
			v.validateObject(currentSubSchema, castCurrentNode, result, context)
			v.validateCommon(currentSubSchema, castCurrentNode, result, context)

			for _, pSchema := range currentSubSchema.propertiesChildren {
				nextNode, ok := castCurrentNode[pSchema.property]
				if ok {
					subContext := NewJsonContext(pSchema.property, context)
					*fieldPath = append(*fieldPath, pSchema.property)
					v.validateRecursive(pSchema, nextNode, result, subContext, fieldPath)
					*fieldPath = (*fieldPath)[:len(*fieldPath)-1]
				}
			}

		// Simple JSON values : string, number, boolean
		case reflect.Bool:
			if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_BOOL, TYPE_BOOLEAN) {
				return
			}
			value := currentNode.(bool)

			v.validateSchema(currentSubSchema, value, result, context)
			v.validateNumber(currentSubSchema, value, result, context)
			v.validateCommon(currentSubSchema, value, result, context)
			v.validateString(currentSubSchema, value, result, context)

		case reflect.String:

			var value interface{}
			switch cn := currentNode.(type) {

			case string:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_STRING) {
					return
				}
				value = cn
				v.validateString(currentSubSchema, value, result, context)
			}
			v.validateSchema(currentSubSchema, value, result, context)
			v.validateNumber(currentSubSchema, value, result, context)
			v.validateCommon(currentSubSchema, value, result, context)
		case reflect.Struct:

			var value interface{}
			switch cn := currentNode.(type) {

			case time.Time:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_DATE) {
					return
				}
				value = cn
			case primitive.Timestamp:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_TIMESTAMP) {
					return
				}
				value = cn
			case primitive.Regex:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_REGEX) {
					return
				}
				value = cn
			case primitive.Decimal128:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_DECIMAL128, TYPE_NUMBER) {
					return
				}
				value = cn
			}

			v.validateSchema(currentSubSchema, value, result, context)
			v.validateNumber(currentSubSchema, value, result, context)
			v.validateCommon(currentSubSchema, value, result, context)
			v.validateString(currentSubSchema, value, result, context)
		case reflect.Array:
			var value interface{}
			switch cn := currentNode.(type) {

			case primitive.ObjectID:
				if !currentSubSchema.containsValidType(currentNode, result, context, TYPE_OBJECT_ID) {
					return
				}
				value = cn
			}

			v.validateSchema(currentSubSchema, value, result, context)
			v.validateNumber(currentSubSchema, value, result, context)
			v.validateCommon(currentSubSchema, value, result, context)
			v.validateString(currentSubSchema, value, result, context)
		}

	}

	result.incrementScore()
}

func (v *subSchema) containsValidType(currentNode interface{}, result *Result, context *JsonContext, types ...string) bool {
	var validType bool
	for _, t := range types {
		if v.types.Contains(t) || v.bsonTypes.Contains(t) {
			validType = true
		}
	}

	if (v.types.IsTyped() || v.bsonTypes.IsTyped()) && !validType {
		result.addInternalError(
			new(InvalidTypeError),
			context,
			currentNode,
			ErrorDetails{
				"expected": fmt.Sprintf("type: %s, bsonType: %s", v.types.String(), v.bsonTypes.String()),
				"given":    types,
			},
		)
		return false
	}

	return true
}

// Different kinds of validation there, subSchema / common / array / object / string...
func (v *subSchema) validateSchema(currentSubSchema *subSchema, currentNode interface{}, result *Result, context *JsonContext) {

	if internalLogEnabled {
		internalLog("validateSchema %s", context.String())
		internalLog(" %v", currentNode)
	}

	if len(currentSubSchema.anyOf) > 0 {

		validatedAnyOf := false
		var bestValidationResult *Result

		for _, anyOfSchema := range currentSubSchema.anyOf {
			if !validatedAnyOf {
				validationResult := anyOfSchema.subValidateWithContext(currentNode, context)
				validatedAnyOf = validationResult.Valid()

				if !validatedAnyOf && (bestValidationResult == nil || validationResult.score > bestValidationResult.score) {
					bestValidationResult = validationResult
				}
			}
		}
		if !validatedAnyOf {

			result.addInternalError(new(NumberAnyOfError), context, currentNode, ErrorDetails{})

			if bestValidationResult != nil {
				// add error messages of closest matching subSchema as
				// that's probably the one the user was trying to match
				result.mergeErrors(bestValidationResult)
			}
		}
	}

	if len(currentSubSchema.oneOf) > 0 {

		nbValidated := 0
		var bestValidationResult *Result

		for _, oneOfSchema := range currentSubSchema.oneOf {
			validationResult := oneOfSchema.subValidateWithContext(currentNode, context)
			if validationResult.Valid() {
				nbValidated++
			} else if nbValidated == 0 && (bestValidationResult == nil || validationResult.score > bestValidationResult.score) {
				bestValidationResult = validationResult
			}
		}

		if nbValidated != 1 {

			result.addInternalError(new(NumberOneOfError), context, currentNode, ErrorDetails{})

			if nbValidated == 0 {
				// add error messages of closest matching subSchema as
				// that's probably the one the user was trying to match
				result.mergeErrors(bestValidationResult)
			}
		}

	}

	if len(currentSubSchema.allOf) > 0 {
		nbValidated := 0

		for _, allOfSchema := range currentSubSchema.allOf {
			validationResult := allOfSchema.subValidateWithContext(currentNode, context)
			if validationResult.Valid() {
				nbValidated++
			}
			result.mergeErrors(validationResult)
		}

		if nbValidated != len(currentSubSchema.allOf) {
			result.addInternalError(new(NumberAllOfError), context, currentNode, ErrorDetails{})
		}
	}

	if currentSubSchema.not != nil {
		validationResult := currentSubSchema.not.subValidateWithContext(currentNode, context)
		if validationResult.Valid() {
			result.addInternalError(new(NumberNotError), context, currentNode, ErrorDetails{})
		}
	}

	if currentSubSchema.dependencies != nil && len(currentSubSchema.dependencies) > 0 {
		if isKind(currentNode, reflect.Map) {
			dependenciesMap, err := mustBeMap(currentNode)
			if err != nil {
				result.addInternalError(new(InvalidTypeError), context, currentNode, ErrorDetails{"error": err.Error()})
			}
			for elementKey := range dependenciesMap {
				if dependency, ok := currentSubSchema.dependencies[elementKey]; ok {
					switch dependency := dependency.(type) {

					case []string:
						for _, dependOnKey := range dependency {
							if _, dependencyResolved := dependenciesMap[dependOnKey]; !dependencyResolved {
								result.addInternalError(
									new(MissingDependencyError),
									context,
									currentNode,
									ErrorDetails{"dependency": dependOnKey},
								)
							}
						}

					case *subSchema:
						dependency.validateRecursive(dependency, currentNode, result, context, dependency.fieldPath)
					}
				}
			}
		}
	}

	if currentSubSchema.expression != nil {
		if err := currentSubSchema.evaluator.Evaluate(currentSubSchema.expression, *currentSubSchema.fieldPath); err != nil {
			result.addInternalError(new(ValidationError), context, currentNode, ErrorDetails{"error": err})
		}
	}

	if currentSubSchema._if != nil {
		validationResultIf := currentSubSchema._if.subValidateWithContext(currentNode, context)
		if currentSubSchema._then != nil && validationResultIf.Valid() {
			validationResultThen := currentSubSchema._then.subValidateWithContext(currentNode, context)
			if !validationResultThen.Valid() {
				result.addInternalError(new(ConditionThenError), context, currentNode, ErrorDetails{})
				result.mergeErrors(validationResultThen)
			}
		}
		if currentSubSchema._else != nil && !validationResultIf.Valid() {
			validationResultElse := currentSubSchema._else.subValidateWithContext(currentNode, context)
			if !validationResultElse.Valid() {
				result.addInternalError(new(ConditionElseError), context, currentNode, ErrorDetails{})
				result.mergeErrors(validationResultElse)
			}
		}
	}

	result.incrementScore()
}

func (v *subSchema) validateCommon(currentSubSchema *subSchema, value interface{}, result *Result, context *JsonContext) {

	if internalLogEnabled {
		internalLog("validateCommon %s", context.String())
		internalLog(" %v", value)
	}

	// const:
	if currentSubSchema._const != nil {
		vString, err := marshalWithoutNumber(value)
		if err != nil {
			result.addInternalError(new(InternalError), context, value, ErrorDetails{"error": err})
		}
		if *vString != *currentSubSchema._const {
			result.addInternalError(new(ConstError),
				context,
				value,
				ErrorDetails{
					"allowed": *currentSubSchema._const,
				},
			)
		}
	}

	// enum:
	if len(currentSubSchema.enum) > 0 {
		has, err := currentSubSchema.ContainsEnum(value)
		if err != nil {
			result.addInternalError(new(InternalError), context, value, ErrorDetails{"error": err})
		}
		if !has {
			result.addInternalError(
				new(EnumError),
				context,
				value,
				ErrorDetails{
					"allowed": strings.Join(currentSubSchema.enum, ", "),
				},
			)
		}
	}

	result.incrementScore()
}

func (v *subSchema) validateArray(currentSubSchema *subSchema, value []interface{}, result *Result, context *JsonContext) {

	if internalLogEnabled {
		internalLog("validateArray %s", context.String())
		internalLog(" %v", value)
	}

	nbValues := len(value)

	// TODO explain
	if currentSubSchema.itemsChildrenIsSingleSchema {
		for i := range value {
			subContext := NewJsonContext(strconv.Itoa(i), context)
			validationResult := currentSubSchema.itemsChildren[0].subValidateWithContext(value[i], subContext)
			result.mergeErrors(validationResult)
		}
	} else {
		if currentSubSchema.itemsChildren != nil && len(currentSubSchema.itemsChildren) > 0 {

			nbItems := len(currentSubSchema.itemsChildren)

			// while we have both schemas and values, check them against each other
			for i := 0; i != nbItems && i != nbValues; i++ {
				subContext := NewJsonContext(strconv.Itoa(i), context)
				validationResult := currentSubSchema.itemsChildren[i].subValidateWithContext(value[i], subContext)
				result.mergeErrors(validationResult)
			}

			if nbItems < nbValues {
				// we have less schemas than elements in the instance array,
				// but that might be ok if "additionalItems" is specified.

				switch currentSubSchema.additionalItems.(type) {
				case bool:
					if !currentSubSchema.additionalItems.(bool) {
						result.addInternalError(new(ArrayNoAdditionalItemsError), context, value, ErrorDetails{})
					}
				case *subSchema:
					additionalItemSchema := currentSubSchema.additionalItems.(*subSchema)
					for i := nbItems; i != nbValues; i++ {
						subContext := NewJsonContext(strconv.Itoa(i), context)
						validationResult := additionalItemSchema.subValidateWithContext(value[i], subContext)
						result.mergeErrors(validationResult)
					}
				}
			}
		}
	}

	// minItems & maxItems
	if currentSubSchema.minItems != nil {
		if nbValues < int(*currentSubSchema.minItems) {
			result.addInternalError(
				new(ArrayMinItemsError),
				context,
				value,
				ErrorDetails{"min": *currentSubSchema.minItems},
			)
		}
	}
	if currentSubSchema.maxItems != nil {
		if nbValues > int(*currentSubSchema.maxItems) {
			result.addInternalError(
				new(ArrayMaxItemsError),
				context,
				value,
				ErrorDetails{"max": *currentSubSchema.maxItems},
			)
		}
	}

	// uniqueItems:
	if currentSubSchema.uniqueItems {
		var stringifiedItems = make(map[string]int)
		for j, v := range value {
			vString, err := marshalWithoutNumber(v)
			if err != nil {
				result.addInternalError(new(InternalError), context, value, ErrorDetails{"err": err})
			}
			if i, ok := stringifiedItems[*vString]; ok {
				result.addInternalError(
					new(ItemsMustBeUniqueError),
					context,
					value,
					ErrorDetails{"type": TYPE_ARRAY, "i": i, "j": j},
				)
			}
			stringifiedItems[*vString] = j
		}
	}

	// contains:

	if currentSubSchema.contains != nil {
		validatedOne := false
		var bestValidationResult *Result

		for i, v := range value {
			subContext := NewJsonContext(strconv.Itoa(i), context)

			validationResult := currentSubSchema.contains.subValidateWithContext(v, subContext)
			if validationResult.Valid() {
				validatedOne = true
				break
			} else {
				if bestValidationResult == nil || validationResult.score > bestValidationResult.score {
					bestValidationResult = validationResult
				}
			}
		}
		if !validatedOne {
			result.addInternalError(
				new(ArrayContainsError),
				context,
				value,
				ErrorDetails{},
			)
			if bestValidationResult != nil {
				result.mergeErrors(bestValidationResult)
			}
		}
	}

	result.incrementScore()
}

func (v *subSchema) validateObject(currentSubSchema *subSchema, value map[string]interface{}, result *Result, context *JsonContext) {

	if internalLogEnabled {
		internalLog("validateObject %s", context.String())
		internalLog(" %v", value)
	}

	// minProperties & maxProperties:
	if currentSubSchema.minProperties != nil {
		if len(value) < int(*currentSubSchema.minProperties) {
			result.addInternalError(
				new(ArrayMinPropertiesError),
				context,
				value,
				ErrorDetails{"min": *currentSubSchema.minProperties},
			)
		}
	}
	if currentSubSchema.maxProperties != nil {
		if len(value) > int(*currentSubSchema.maxProperties) {
			result.addInternalError(
				new(ArrayMaxPropertiesError),
				context,
				value,
				ErrorDetails{"max": *currentSubSchema.maxProperties},
			)
		}
	}

	// required:
	for _, requiredProperty := range currentSubSchema.required {
		_, ok := value[requiredProperty]
		if ok {
			result.incrementScore()
		} else {
			result.addInternalError(
				new(RequiredError),
				context,
				value,
				ErrorDetails{"property": requiredProperty},
			)
		}
	}

	// additionalProperty & patternProperty:
	if currentSubSchema.additionalProperties != nil {

		switch currentSubSchema.additionalProperties.(type) {
		case bool:

			if !currentSubSchema.additionalProperties.(bool) {

				for pk := range value {

					found := false
					for _, spValue := range currentSubSchema.propertiesChildren {
						if pk == spValue.property {
							found = true
						}
					}

					ppHas, ppMatch := v.validatePatternProperty(currentSubSchema, pk, value[pk], result, context)

					if found {

						if ppHas && !ppMatch {
							result.addInternalError(
								new(AdditionalPropertyNotAllowedError),
								context,
								value[pk],
								ErrorDetails{"property": pk},
							)
						}

					} else {

						if !ppHas || !ppMatch {
							result.addInternalError(
								new(AdditionalPropertyNotAllowedError),
								context,
								value[pk],
								ErrorDetails{"property": pk},
							)
						}

					}
				}
			}

		case *subSchema:

			additionalPropertiesSchema := currentSubSchema.additionalProperties.(*subSchema)
			for pk := range value {

				found := false
				for _, spValue := range currentSubSchema.propertiesChildren {
					if pk == spValue.property {
						found = true
					}
				}

				ppHas, ppMatch := v.validatePatternProperty(currentSubSchema, pk, value[pk], result, context)

				if found {

					if ppHas && !ppMatch {
						validationResult := additionalPropertiesSchema.subValidateWithContext(value[pk], context)
						result.mergeErrors(validationResult)
					}

				} else {

					if !ppHas || !ppMatch {
						validationResult := additionalPropertiesSchema.subValidateWithContext(value[pk], context)
						result.mergeErrors(validationResult)
					}

				}

			}
		}
	} else {

		for pk := range value {

			ppHas, ppMatch := v.validatePatternProperty(currentSubSchema, pk, value[pk], result, context)

			if ppHas && !ppMatch {

				result.addInternalError(
					new(InvalidPropertyPatternError),
					context,
					value[pk],
					ErrorDetails{
						"property": pk,
						"pattern":  currentSubSchema.PatternPropertiesString(),
					},
				)
			}

		}
	}

	// propertyNames:
	if currentSubSchema.propertyNames != nil {
		for pk := range value {
			validationResult := currentSubSchema.propertyNames.subValidateWithContext(pk, context)
			if !validationResult.Valid() {
				result.addInternalError(new(InvalidPropertyNameError),
					context,
					value, ErrorDetails{
						"property": pk,
					})
				result.mergeErrors(validationResult)
			}
		}
	}

	result.incrementScore()
}

func (v *subSchema) validatePatternProperty(currentSubSchema *subSchema, key string, value interface{}, result *Result, context *JsonContext) (has bool, matched bool) {

	if internalLogEnabled {
		internalLog("validatePatternProperty %s", context.String())
		internalLog(" %s %v", key, value)
	}

	has = false

	validatedkey := false

	for pk, pv := range currentSubSchema.patternProperties {
		if matches, _ := regexp.MatchString(pk, key); matches {
			has = true
			subContext := NewJsonContext(key, context)
			validationResult := pv.subValidateWithContext(value, subContext)
			result.mergeErrors(validationResult)
			validatedkey = true
		}
	}

	if !validatedkey {
		return has, false
	}

	result.incrementScore()

	return has, true
}

func (v *subSchema) validateString(currentSubSchema *subSchema, value interface{}, result *Result, context *JsonContext) {

	// Ignore JSON numbers
	if isJSONNumber(value) {
		return
	}

	// Ignore non strings
	if !isKind(value, reflect.String) {
		return
	}

	if internalLogEnabled {
		internalLog("validateString %s", context.String())
		internalLog(" %v", value)
	}

	stringValue := value.(string)

	// minLength & maxLength:
	if currentSubSchema.minLength != nil {
		if utf8.RuneCount([]byte(stringValue)) < int(*currentSubSchema.minLength) {
			result.addInternalError(
				new(StringLengthGTEError),
				context,
				value,
				ErrorDetails{"min": *currentSubSchema.minLength},
			)
		}
	}
	if currentSubSchema.maxLength != nil {
		if utf8.RuneCount([]byte(stringValue)) > int(*currentSubSchema.maxLength) {
			result.addInternalError(
				new(StringLengthLTEError),
				context,
				value,
				ErrorDetails{"max": *currentSubSchema.maxLength},
			)
		}
	}

	// pattern:
	if currentSubSchema.pattern != nil {
		if !currentSubSchema.pattern.MatchString(stringValue) {
			result.addInternalError(
				new(DoesNotMatchPatternError),
				context,
				value,
				ErrorDetails{"pattern": currentSubSchema.pattern},
			)

		}
	}

	// format
	if currentSubSchema.format != "" {
		if !FormatCheckers.IsFormat(currentSubSchema.format, stringValue) {
			result.addInternalError(
				new(DoesNotMatchFormatError),
				context,
				value,
				ErrorDetails{"format": currentSubSchema.format},
			)
		}
	}

	result.incrementScore()
}

func (v *subSchema) validateNumber(currentSubSchema *subSchema, value interface{}, result *Result, context *JsonContext) {

	// Ignore non numbers
	if !isJSONNumber(value) && !isNumber(value) {
		return
	}

	if internalLogEnabled {
		internalLog("validateNumber %s", context.String())
		internalLog(" %v", value)
	}

	float64Value := mustBeNumber(value)

	// multipleOf:
	if currentSubSchema.multipleOf != nil {
		if q := new(big.Rat).Quo(float64Value, currentSubSchema.multipleOf); !q.IsInt() {
			result.addInternalError(
				new(MultipleOfError),
				context,
				fmt.Sprintf("%v", float64Value),
				ErrorDetails{"multiple": currentSubSchema.multipleOf},
			)
		}
	}

	//maximum & exclusiveMaximum:
	if currentSubSchema.maximum != nil {
		if float64Value.Cmp(currentSubSchema.maximum) == 1 {
			result.addInternalError(
				new(NumberLTEError),
				context,
				fmt.Sprintf("%v", float64Value),
				ErrorDetails{
					"max": currentSubSchema.maximum,
				},
			)
		}
	}
	if currentSubSchema.exclusiveMaximum != nil {
		if float64Value.Cmp(currentSubSchema.exclusiveMaximum) >= 0 {
			result.addInternalError(
				new(NumberLTError),
				context,
				fmt.Sprintf("%v", float64Value),
				ErrorDetails{
					"max": currentSubSchema.maximum,
				},
			)
		}
	}

	//minimum & exclusiveMinimum:
	if currentSubSchema.minimum != nil {
		if float64Value.Cmp(currentSubSchema.minimum) == -1 {
			result.addInternalError(
				new(NumberGTEError),
				context,
				fmt.Sprintf("%v", float64Value),
				ErrorDetails{
					"min": currentSubSchema.minimum,
				},
			)
		}
	}
	if currentSubSchema.exclusiveMinimum != nil {
		if float64Value.Cmp(currentSubSchema.exclusiveMinimum) <= 0 {
			result.addInternalError(
				new(NumberGTError),
				context,
				fmt.Sprintf("%v", float64Value),
				ErrorDetails{
					"min": currentSubSchema.minimum,
				},
			)
		}
	}

	// format
	if currentSubSchema.format != "" {
		if !FormatCheckers.IsFormat(currentSubSchema.format, float64Value) {
			result.addInternalError(
				new(DoesNotMatchFormatError),
				context,
				value,
				ErrorDetails{"format": currentSubSchema.format},
			)
		}
	}

	result.incrementScore()
}
