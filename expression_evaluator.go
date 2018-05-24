package gojsonschema

type ExpressionEvaluator interface {
	Evaluate(expression interface{}, fieldPath []string) error
}

type NoopEvaluator struct {}

func NewNoopEvaluator() *NoopEvaluator {
	return &NoopEvaluator{}
}

func (evaluator *NoopEvaluator) Evaluate(expression interface{}, fieldPath []string) error {
	return nil
}
