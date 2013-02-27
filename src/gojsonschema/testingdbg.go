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

	for i := range s.children {
		debugDisplayJsonSchemaRecursive(s.children[i], level+1)
	}

}
