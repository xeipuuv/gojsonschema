// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      27-02-2013

package gojsonschema

import (
	"gojsonreference"
)

type JsonSchema struct {
	id          *string
	title       *string
	description *string
	types       JsonSchemaType

	ref *gojsonreference.JsonReference

	definitionsChildren []*JsonSchema
	itemsChild          *JsonSchema
	propertiesChildren  []*JsonSchema

	parent *JsonSchema

	property string

	schema *gojsonreference.JsonReference

	// validation : number / integer
	multipleOf       *float64
	
	maximum          *float64
	exclusiveMaximum bool
}

func (s *JsonSchema) AddDefinitionChild(child *JsonSchema) {
	s.definitionsChildren = append(s.definitionsChildren, child)
}

func (s *JsonSchema) SetItemsChild(child *JsonSchema) {
	s.itemsChild = child
}

func (s *JsonSchema) AddPropertiesChild(child *JsonSchema) {
	s.propertiesChildren = append(s.propertiesChildren, child)
}
