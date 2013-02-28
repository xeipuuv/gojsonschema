package gojsonschema

import (
	"fmt"
)

func DebugDisplayJsonSchema(d *JsonSchemaDocument) {
	debugDisplayJsonSchemaRecursive(d.rootSchema, 0)
}

func debugDisplayJsonSchemaRecursive(s *JsonSchema, level int) {
	for i := 0; i != level; i++ {
		fmt.Printf(" ")
	}

	fmt.Printf(s.property)

	if s.ref != nil {
		fmt.Printf(" | ref %s", s.ref.String())
	}

	if s.id != nil {
		fmt.Printf(" | id %s", *s.id)
	}

	fmt.Printf(" | type %s", s.etype)

	fmt.Printf("\n")

	for i := range s.definitionsChildren {
		debugDisplayJsonSchemaRecursive(s.definitionsChildren[i], level+1)
	}

	for i := range s.propertiesChildren {
		debugDisplayJsonSchemaRecursive(s.propertiesChildren[i], level+1)
	}

	if s.itemsChild != nil {
		debugDisplayJsonSchemaRecursive(s.itemsChild, level+1)
	}

}
