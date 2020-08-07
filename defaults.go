package gojsonschema

import (
	"errors"
	"fmt"
)

// Author: Robert Kotcher
//		   robert@synthesis.ai
//         github.com/rkotcher-synthesisai

// NOTE special thanks to GH user juju for the inspiration behind this functionality
// I studied their fork of this repo when I was working on this. Some notable differences
// are: 1) InsertDefaults does not accept nil now, 2) I added support for default
// values w/in arrays. You can check out the work juju did on their fork at:
// https://godoc.org/github.com/juju/gojsonschema

// InsertDefaults takes a generic interface (because it could either be an
// object or an array, and attemps to fill it with as many defaults as possible
func (s *Schema) InsertDefaults(into interface{}) (m interface{}, returnErr error) {
	defer panicHandler(&returnErr)

	// We need to get the outermost document before entering the recursive function
	// because we'll recurse down into this map as well.
	schemaMap := s.getDocumentMap()

	insertRecursively(into, schemaMap)

	return into, nil // err is filled if panic
}

func panicHandler(err *error) {
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

func (s *Schema) getDocumentMap() map[string]interface{} {
	f, _ := s.pool.GetDocument(s.documentReference)
	return f.Document.(map[string]interface{})
}

// insertRecursively inserts into "into", which is either an array or an object
func insertRecursively(into interface{}, from map[string]interface{}) {

	switch t := from["type"]; t {

	case "array":
		// intoAsArray represents many objects of the same type
		intoAsArray := into.([]map[string]interface{})

		// nextMap represents the subSchema that we want each item in the array
		// to conform to
		nextMap := from["items"].(map[string]interface{})

		for _, example := range intoAsArray {
			insertRecursively(example, nextMap)
		}

	case "object":
		intoAsObject := into.(map[string]interface{})

		// nextMap represents the subSchema that we want this single item to
		// conform to
		properties := from["properties"].(map[string]interface{})

		for property, _nextSchema := range properties {

			nextSchema := _nextSchema.(map[string]interface{})

			// This block becomes active if stuff already exists in "into"
			// for this property
			if v, ok := intoAsObject[property]; ok {
				switch v.(type) {

				case map[string]interface{}:
					if innerMapAsObj, ok := v.(map[string]interface{}); ok {
						insertRecursively(innerMapAsObj, nextSchema)
					}

				case []map[string]interface{}:
					if innerMapAsArr, ok := v.([]map[string]interface{}); ok {
						insertRecursively(innerMapAsArr, nextSchema)
					}
				}
				continue
			}

			// We can't step deeper so we're at an actual key/value
			// Check to see if we should add a default
			if d, ok := nextSchema["default"]; ok {
				intoAsObject[property] = d
				continue
			}

			// Finally, if the next schema does exists but there is nothing
			// in the input object, we want to create a temporary interface, just
			// in case a nested object has defaults
			if t, ok := nextSchema["type"]; ok {
				switch t {

				// If there's an array here, we want to initialize it to an empty
				// array and leave it at that.
				case "array":
					emptyarr := make([]map[string]interface{}, 0)
					intoAsObject[property] = emptyarr

				case "object":
					tmpTarget := make(map[string]interface{})
					insertRecursively(tmpTarget, nextSchema)
					if len(tmpTarget) > 0 {
						intoAsObject[property] = tmpTarget
					}
				}
			}
		}
	}
}
