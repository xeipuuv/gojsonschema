package gojsonschema

import (
	"reflect"
	"regexp"
)

type (
	// FormatChecker is the interface all formatters added to FormatterChain must implement
	FormatChecker interface {
		IsFormat(input string) bool
	}

	// FormatterChain holds the formatters
	FormatterChain struct {
		formatters map[string]FormatChecker
	}

	// EmailFormatter verifies emails
	EmailFormatter struct{}
)

var (
	// Formatters holds the valid formatters, and is a public variable
	// so library users can add custom formatters
	Formatters = FormatterChain{
		formatters: map[string]FormatChecker{
			"email": EmailFormatter{},
		},
	}

	// Regex credit: https://github.com/asaskevich/govalidator
	rxEmail = regexp.MustCompile("^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")
)

// Add adds a FormatChecker to the FormatterChain
// The name used will be the value used for the format key in your json schema
func (c *FormatterChain) Add(name string, f FormatChecker) *FormatterChain {
	c.formatters[name] = f

	return c
}

// Remove deletes a FormatChecker from the FormatterChain (if it exists)
func (c *FormatterChain) Remove(name string) *FormatterChain {
	delete(c.formatters, name)

	return c
}

// Has checks to see if the FormatterChain holds a FormatChecker with the given name
func (c *FormatterChain) Has(name string) bool {
	_, ok := c.formatters[name]

	return ok
}

// IsFormat will check an input against a FormatChecker with the given name
// to see if it is the correct format
func (c *FormatterChain) IsFormat(name string, input interface{}) bool {
	f, ok := c.formatters[name]

	if !ok {
		return false
	}

	if !isKind(input, reflect.String) {
		return false
	}

	inputString := input.(string)

	return f.IsFormat(inputString)
}

func (f EmailFormatter) IsFormat(input string) bool {
	return rxEmail.MatchString(input)
}
