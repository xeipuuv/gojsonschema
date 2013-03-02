// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Defines the structure of a schema. 
//					A schema can have sub-schemas.
// 
// created			27-02-2013

package gojsonschema

import (
	"errors"
	"gojsonreference"
	"regexp"
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
	minimum          *float64
	exclusiveMinimum bool

	// validation : string
	minLength *int
	maxLength *int
	pattern   *regexp.Regexp

	// validation : object
	minProperties *int
	maxProperties *int

	required []string

	// validation : array
	minItems    *int
	maxItems    *int
	uniqueItems bool

	// validation : all
	enum []string

	// validation : schema
	oneOf []*JsonSchema
	anyOf []*JsonSchema
	allOf []*JsonSchema
	not   *JsonSchema
}

func (s *JsonSchema) AddEnum(i interface{}) error {

	is, err := marshalToString(i)
	if err != nil {
		return err
	}

	if isStringInSlice(s.enum, *is) {
		return errors.New("enum items must be unique")
	}

	s.enum = append(s.enum, *is)

	return nil
}

func (s *JsonSchema) AddOneOf(schema *JsonSchema) {
	s.oneOf = append(s.oneOf, schema)
}

func (s *JsonSchema) AddAllOf(schema *JsonSchema) {
	s.allOf = append(s.allOf, schema)
}

func (s *JsonSchema) AddAnyOf(schema *JsonSchema) {
	s.anyOf = append(s.anyOf, schema)
}

func (s *JsonSchema) SetNot(schema *JsonSchema) {
	s.not = schema
}

func (s *JsonSchema) HasEnum(i interface{}) (bool, error) {

	is, err := marshalToString(i)
	if err != nil {
		return false, err
	}

	return isStringInSlice(s.enum, *is), nil
}

func (s *JsonSchema) AddRequired(value string) error {

	if isStringInSlice(s.required, value) {
		return errors.New("required items must be unique")
	}

	s.required = append(s.required, value)

	return nil
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

func (s *JsonSchema) HasProperty(name string) bool {

	for _, v := range s.propertiesChildren {
		if v.property == name {
			return true
		}
	}
	return false
}
