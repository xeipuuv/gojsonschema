package gojsonschema

type (

	// FormatCheckerWithError exposes a new interface for IsFormat signature. Ideally, there should be ResultError returned along with IsFmt so users can easily
	// customize the message. For now, this is a proof of concept. If met with acceptance, then we can slowly deprecate to new definition.
	FormatCheckerWithError interface {
		FormatChecker
		IsFormatWithError(input interface{}) (bool, ResultError)
	}

	baseFormatCheckerWithError struct {
		FormatChecker
	}
)

var _ FormatCheckerWithError = (*baseFormatCheckerWithError)(nil)

// convertToNewChecker
func convertToNewChecker(oldChecker FormatChecker) FormatCheckerWithError {
	newChecker := &baseFormatCheckerWithError{oldChecker}
	return newChecker
}

//IsFormatWithError returns whether format is met and corresponding result error. Extends existing FormatCheckers.
func (b *baseFormatCheckerWithError) IsFormatWithError(input interface{}) (bool, ResultError) {
	isFmt := b.IsFormat(input)
	if isFmt {
		return true, nil
	}

	return false, new(DoesNotMatchFormatError)
}

// AddCheckerWithError extends FormatCheckerChain to add this new FormatChecker interface.
func (c *FormatCheckerChain) AddCheckerWithError(name string, f FormatCheckerWithError) *FormatCheckerChain {
	lock.Lock()
	c.formatters[name] = f
	lock.Unlock()
	return c
}

// IsFormatWithError will check if an input matches corresponding format and returns appropriate error associated to it.
func (c *FormatCheckerChain) IsFormatWithError(name string, input interface{}) (bool, ResultError) {
	lock.RLock()
	f, ok := c.formatters[name]
	lock.RUnlock()

	// If a format is unrecognized it should always pass validation
	if !ok {
		return true, nil
	}

	return f.IsFormatWithError(input)
}
