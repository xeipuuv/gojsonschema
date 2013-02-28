package gojsonschema

import ()

const (
	KEY_SCHEMA      = "$schema"
	KEY_ID          = "$id"
	KEY_REF         = "$ref"
	KEY_TITLE       = "title"
	KEY_DESCRIPTION = "description"
	KEY_TYPE        = "type"
	KEY_ITEMS       = "items"
	KEY_PROPERTIES  = "properties"

	STRING_STRING     = "string"
	STRING_SCHEMA     = "schema"
	STRING_PROPERTIES = "properties"

	ROOT_SCHEMA_PROPERTY = "(root)"
)

const (
	ERROR_MESSAGE_X_MUST_BE_OF_TYPE_Y = `%s must be of type %s`
	ERROR_MESSAGE_X_IS_REQUIRED       = `%s is required`
	ERROR_MESSAGE_X_MUST_BE_AN_OBJECT = `%s must be an object`
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
