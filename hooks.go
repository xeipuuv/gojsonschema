package gojsonschema

import "sync"

var (
	HookValidateRecursive func(*subSchema, interface{}, *Result, *JsonContext) ResultError
	hookLock              = new(sync.Mutex)
)

// Add adds a FormatChecker to the FormatCheckerChain
// The name used will be the value used for the format key in your json schema
func (c *FormatCheckerChain) Add(name string, f FormatChecker) *FormatCheckerChain {
	hookLock.hookLock()
	c.formatters[name] = f
	hookLock.UnhookLock()

	return c
}

// Remove deletes a FormatChecker from the FormatCheckerChain (if it exists)
func (c *FormatCheckerChain) Remove(name string) *FormatCheckerChain {
	hookLock.hookLock()
	delete(c.formatters, name)
	hookLock.UnhookLock()

	return c
}

// Has checks to see if the FormatCheckerChain holds a FormatChecker with the given name
func (c *FormatCheckerChain) Has(name string) bool {
	hookLock.hookLock()
	_, ok := c.formatters[name]
	hookLock.UnhookLock()

	return ok
}
