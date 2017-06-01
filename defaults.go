// Author: Bodie Solomon
//         bodie@synapsegarden.net
//         github.com/binary132
//
//         2015-02-16

package gojsonschema

import (
	"errors"
	"fmt"
)

// InsertDefaults takes a map[string]interface{} and inserts any missing
// default values specified in the JSON-Schema document.  If the given "into"
// map is nil, it will be created.
//
// This is an initial implementation which does not support refs, etc.
// The schema must be for an object, not a bare value.
func (s *Schema) InsertDefaults(into map[string]interface{}) (m map[string]interface{}, returnErr error) {
	// Handle any panics caused by malformed schemas.
	defer schemaPanicHandler(&returnErr)

	properties := s.getDocProperties()

	// Make sure the "into" map isn't nil.
	if into == nil {
		into = make(map[string]interface{})
	}

	// Now insert the default values at the keys in the "into" map,
	// non-destructively.
	iterateAndInsert(into, properties)

	return into, nil
}

// schemaPanics catches panics caused by type assertions or gets on schemas
// which have somehow slipped past validation.
func schemaPanicHandler(err *error) {
	if r := recover(); r != nil {
		var msg string
		switch t := r.(type) {
		case error:
			msg = fmt.Sprintf("schema error caused a panic: %s", t.Error())
		default:
			msg = fmt.Sprintf("unknown panic: %#v", t)
		}
		*err = errors.New(msg)
	}
}

// getDocProperties retrieves the "properties" key of the standalone doc of
// the schema.
//
// Malformed schemas cause a panic.
func (s *Schema) getDocProperties() map[string]interface{} {
	f := s.pool.GetStandaloneDocument()

	// We need a map[string]interface because we have to check for a
	// default key.
	docMap := f.(map[string]interface{})

	// We need a schema with properties, since this feature does not
	// support raw value schemas.
	m := docMap["properties"]

	return m.(map[string]interface{})
}

// iterateAndInsert takes a target map and inserts any missing default values
// as specified in the properties map, according to JSON-Schema.
//
// Malformed schemas cause a panic.
func iterateAndInsert(into, properties map[string]interface{}) {
	for property, schema := range properties {
		// Iterate over each key of the schema.  Each key should have
		// a schema describing the key's properties.  If it does not,
		// ignore this key.
		typedSchema := schema.(map[string]interface{})

		if v, ok := into[property]; ok {
			// If there is already a map value in the target for
			// this key, don't overwrite it; step in.  Ignore
			// non-map values.
			if innerMap, ok := v.(map[string]interface{}); ok {
				// The schema should have an inner
				// schema since it's an object
				schemaProperties := typedSchema["properties"]
				typedProperties := schemaProperties.(map[string]interface{})
				iterateAndInsert(innerMap, typedProperties)
			}
			continue
		}

		if d, ok := typedSchema["default"]; ok {
			// Most basic case: we have a default value. Done for
			// this key.
			into[property] = d
			continue
		}

		if p, ok := typedSchema["properties"]; ok {
			// If we have a "properties" key, this is an object,
			// and we need to go deeper.
			innerSchema := p.(map[string]interface{})
			tmpTarget := make(map[string]interface{})
			iterateAndInsert(tmpTarget, innerSchema)
			if len(tmpTarget) > 0 {
				into[property] = tmpTarget
			}
		}
	}
}
