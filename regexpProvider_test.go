// Copyright 2018 johandorland ( https://github.com/johandorland )
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gojsonschema

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type customCompiled struct {
	pattern           *regexp.Regexp
	expr              string
	matchStringCalled int
}

func (cc *customCompiled) MatchString(s string) bool {
	cc.matchStringCalled++
	return cc.pattern.MatchString(s)
}

type customRegexpProvider struct {
	compileCalled  int
	compiledRegexp map[string]*customCompiled
}

func (c *customRegexpProvider) Compile(expr string) (CompiledRegexp, error) {
	c.compileCalled++
	pattern, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	cc := &customCompiled{
		pattern: pattern,
		expr:    expr,
	}
	if c.compiledRegexp == nil {
		c.compiledRegexp = make(map[string]*customCompiled)
	}
	c.compiledRegexp[expr] = cc
	return cc, nil
}

func TestCustomRegexpProvider(t *testing.T) {
	// Verify that the RegexpProvider is used
	loader := NewStringLoader(`{
			"patternProperties": {
				"f.*o": {"type": "integer"},
				"b.*r": {"type": "string", "pattern": "^a*$"}
			}
		}`)

	sl := NewSchemaLoader()
	customRegexpProvider := &customRegexpProvider{}
	sl.RegexpProvider = customRegexpProvider
	d, err := sl.Compile(loader)
	assert.Nil(t, err)
	assert.NotNil(t, d.regexp)

	loader = NewStringLoader(`{"foo": 1, "foooooo" : 2, "bar": "a", "baaaar": "aaaa"}`)
	r, err := d.Validate(loader)
	assert.Nil(t, err)
	assert.Empty(t, r.errors)
	assert.Equal(t, 3, customRegexpProvider.compileCalled)
	assert.Equal(t, 4, customRegexpProvider.compiledRegexp["f.*o"].matchStringCalled)
	assert.Equal(t, 4, customRegexpProvider.compiledRegexp["b.*r"].matchStringCalled)
	assert.Equal(t, 2, customRegexpProvider.compiledRegexp["^a*$"].matchStringCalled)
}
