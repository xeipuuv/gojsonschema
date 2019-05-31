package gojsonschema

type Hook func(*SubSchema, interface{}, *Result, *JsonContext) (error)

func (s *Schema) WithHooks(h []Hook) *Schema {
	c := *s
	c.hooks = h
	return c
}

func (s *Schema) Hooks() []Hook {
	return s.hooks
}
