// @author       sigu-399
// @description  An implementation of JSON Schema, draft v4 - Go language
// @created      27-02-2013

package gojsonschema

import ()

type JsonSchema struct {
	id          *string
	title       *string
	description *string	

	ref *string

	children []*JsonSchema

	property *string
}

func (s *JsonSchema) AddChild(child *JsonSchema) {
	s.children = append(s.children, child)
}
