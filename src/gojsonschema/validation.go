// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      28-02-2013

package gojsonschema

import ()

type ValidationResult struct {
	valid bool
}

func (v *ValidationResult) IsValid() bool {
	return v.valid
}

func Validate(document interface{}, validator *JsonSchemaDocument) ValidationResult {

	return ValidationResult{valid: true}
}
