package gojsonschema

import ()

const (
	ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y = `%s must be of type %s`
	ERROR_MESSAGE_X_IS_REQUIRED       = `%s is required`
)

var JSON_TYPES []string
var SCHEMA_TYPES []string

func init() {
	JSON_TYPES = []string{
		`array`,
		`boolean`,
		`integer`,
		`number`,
		`null`,
		`object`,
		`string`}
	SCHEMA_TYPES = []string{
		`array`,
		`boolean`,
		`integer`,
		`number`,
		`object`,
		`string`}
}
