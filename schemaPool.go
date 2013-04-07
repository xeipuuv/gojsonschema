// Copyright 2013 sigu-399 ( https://github.com/sigu-399 )
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

// author           sigu-399
// author-github    https://github.com/sigu-399
// author-mail      sigu.399@gmail.com
// 
// repository-name  gojsonschema
// repository-desc  An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Defines resources pooling. 
//                  Eases referencing and avoids downloading the same resource twice.			
// 
// created          26-02-2013

package gojsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sigu-399/gojsonreference"
	"io/ioutil"
	"net/http"
	"strings"
)

type schemaPool struct {
	schemaPoolDocuments map[string]*schemaPoolDocument
}

func newSchemaPool() *schemaPool {
	p := &schemaPool{}
	p.schemaPoolDocuments = make(map[string]*schemaPoolDocument)
	return p
}

func (p *schemaPool) GetPoolDocument(reference gojsonreference.JsonReference) (*schemaPoolDocument, error) {

	var err error

	if !reference.IsCanonical() {
		return nil, errors.New(fmt.Sprintf("Reference must be canonical %s", reference))
	}

	refToUrl := reference
	refToUrl.GetUrl().Fragment = ""

	var spd *schemaPoolDocument

	for k := range p.schemaPoolDocuments {
		if k == refToUrl.String() {
			spd = p.schemaPoolDocuments[k]
			fmt.Printf("Found in pool %s\n", refToUrl.String())
		}
	}

	if spd != nil {
		return spd, nil
	}

	var document interface{}

	if reference.HasFileScheme {

		filename := strings.Replace(refToUrl.String(), "file://", "")
		document, err = GetFileJson(filename)
		if err != nil {
			return nil, err
		}

	} else {

		document, err = GetHttpJson(refToUrl.String())
		if err != nil {
			return nil, err
		}

	}

	spd = &schemaPoolDocument{Document: document}
	p.schemaPoolDocuments[refToUrl.String()] = spd

	fmt.Printf("Added to pool %s\n", refToUrl.String())

	return spd, nil
}

type schemaPoolDocument struct {
	Document interface{}
}

func GetHttpJson(url string) (interface{}, error) {

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Could not access schema " + resp.Status)
	}

	bodyBuff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var document interface{}
	err = json.Unmarshal(bodyBuff, &document)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func GetFileJson(filename string) (interface{}, error) {

	bodyBuff, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var document interface{}
	err = json.Unmarshal(bodyBuff, &document)
	if err != nil {
		return nil, err
	}

	return document, nil
}
