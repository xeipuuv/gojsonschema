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

	if s.property == nil {
		fmt.Printf("(nil)")
	} else {
		fmt.Printf(*s.property)
	}

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
