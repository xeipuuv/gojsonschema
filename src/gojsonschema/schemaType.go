// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Helper structure to handle schema types, and the combination of them.			
// 
// created      	28-02-2013

package gojsonschema

import (
	"errors"
	"fmt"
	"strings"
)

type JsonSchemaType struct {
	types []string
}

func (t *JsonSchemaType) Add(etype string) error {

	if !isStringInSlice(JSON_TYPES, etype) {
		return errors.New(fmt.Sprintf("%s is not a valid type", etype))
	}

	if t.HasType(etype) {
		return errors.New(fmt.Sprintf("%s type is duplicated", etype))
	}

	t.types = append(t.types, etype)

	return nil
}

func (t *JsonSchemaType) HasType(etype string) bool {

	for _, v := range t.types {
		if v == etype {
			return true
		}
	}

	return false
}

func (t *JsonSchemaType) String() string {
	return strings.Join(t.types, ",")
}
