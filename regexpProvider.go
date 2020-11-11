package gojsonschema

import (
	"regexp"
)

var (
	defaultRegexProvider = golangRegexpProvider{}
)

// RegexpProvider An interface to a regex implementation
type RegexpProvider interface {
	// Compile Compiles an expression and returns a CompiledRegexp
	Compile(expr string) (CompiledRegexp, error)
}

// CompiledRegexp A compiled expression
type CompiledRegexp interface {
	// MatchString Tests if the string matches the compiled expression
	MatchString(s string) bool
}

type golangRegexpProvider struct {
}

func (golangRegexpProvider) Compile(expr string) (CompiledRegexp, error) {
	return regexp.Compile(expr)
}

func getDefaultRegexpProvider() RegexpProvider {
	return defaultRegexProvider
}
