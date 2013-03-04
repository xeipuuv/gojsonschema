// author  			sigu-399
// author-github 	https://github.com/sigu-399
// author-mail		sigu.399@gmail.com
// 
// repository-name	gojsonschema
// repository-desc 	An implementation of JSON Schema, based on IETF's draft v4 - Go language.
// 
// description		Defines resources pooling. 
//					Eases referencing and avoids downloading the same resource twice.			
// 
// created      	26-02-2013

package gojsonschema

import (
	"encoding/json"
	"errors"
	"fmt"
	"gojsonreference"
	"io/ioutil"
	"net/http"
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

	if !reference.HasFullUrl {
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

	document, err := GetHttpJson(refToUrl.String())
	if err != nil {
		return nil, err
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
