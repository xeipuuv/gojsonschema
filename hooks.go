package gojsonschema

type Hook func(*SubSchema, interface{}, *Result, *JsonContext) error

func (s *Schema) SetHooks(h []Hook) {
	s.hooks = h
}

func (s *Schema) Hooks() []Hook {
	return s.hooks
}
