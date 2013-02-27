package gojsonschema

import ()

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
