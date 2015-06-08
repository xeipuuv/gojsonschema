package goswagger

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/apcera/gojsonschema"
)

type Schema struct {
	Swagger             string                 `json:"swagger"`
	Host                string                 `json:"host"`
	BasePath            string                 `json:"basePath"`
	Info                map[string]string      `json:"info"`
	Produces            []string               `json:"produces"`
	Paths               map[string]Path        `json:"paths"`
	Definitions         map[string]interface{} `json:"definitions"`
	Parameters          interface{}            `json:"parameters"`
	Responses           interface{}            `json:"responses"`
	SecurityDefinitions interface{}            `json:"securityDefinitions"`
	Security            interface{}            `json:"security"`
	Tags                interface{}            `json:"tags"`
	ExternalDocs        interface{}            `json:"externalDocs"`
}

type Path struct {
	Ref        string      `json:"ref"`
	Get        Operation   `json:"get"`
	Put        Operation   `json:"put"`
	Post       Operation   `json:"post"`
	Delete     Operation   `json:"delete"`
	Options    Operation   `json:"options"`
	Head       Operation   `json:"head"`
	Patch      Operation   `json:"patch"`
	Parameters []Parameter `json:"parameters"`
}

type Operation struct {
	Tags         []string    `json:"tags"`
	Summmary     string      `json:"summary"`
	Description  string      `json:"description"`
	ExternalDocs interface{} `json:"externalDocs"`
	OperationId  string      `json:"operationId"`
	Consumes     []string    `json:"consumes"`
	Produces     []string    `json:"produces"`
	Parameters   []Parameter `json:"parameters"`
	Responses    interface{} `json:"responses"`
	Schemes      []string    `json:"schemes"`
	Deprecated   bool        `json:"deprecated"`
	Security     interface{} `json:"security"`
}

type Parameter struct {
	Name        string                 `json:"name"`
	In          string                 `json:"in"`
	Description string                 `json:"description"`
	Required    bool                   `json:"required"`
	SchemaDef   map[string]interface{} `json:"schema"`
}

// Validates a given request against the schema defined for that path/operation in the Swagger spec.
// Currently only validates the body parameters; does not validate query strings or path parameters.
func (swag *Schema) ValidateHTTPRequest(req *http.Request) (*gojsonschema.Result, error) {
	result := &gojsonschema.Result{}

	thisPath, err := GetMatchingPath(swag.Paths, req.URL.Path)
	if err != nil {
		return nil, err
	}

	var op Operation
	switch req.Method {
	case "GET":
		op = thisPath.Get
	case "POST":
		op = thisPath.Post
	case "PUT":
		op = thisPath.Put
	case "DELETE":
		op = thisPath.Delete
	case "PATCH":
		op = thisPath.Patch
	case "OPTIONS":
		op = thisPath.Options
	case "HEAD":
		op = thisPath.Head
	default:
		panic(errors.New("Unsupported HTTP operation"))
	}

	if op.Responses == nil {
		return nil, errors.New(fmt.Sprintf("The %s operation is not defined on path %s", req.Method, req.URL.Path))
	}

	for _, p := range op.Parameters {
		if p.In == "body" {
			// Load body into schema; delayed until now in case it was unneeded
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			defer req.Body.Close()

			//fmt.Println("schemaDef is:", p.SchemaDef)
			ForceCanonicalRefs(swag, p.SchemaDef)
			//fmt.Println("schemaDef is:", p.SchemaDef)
			// Validate
			return ValidatePartialSchema(swag, p.SchemaDef, string(body))
			_ = body
		}
	}

	// Did validate against anthing; reconsider how to handle this
	return result, nil
}

// Chooses the swagger path corresponding to the requested url; resolves path parameters
// in a simple way that isn't very robust.  It doesnt do any validation on the provided {var}
// even though the path contains schema information for it.
func GetMatchingPath(pathMap map[string]Path, reqPath string) (*Path, error) {
	//Simple implentation; if path is a key then dont waste time with regex testing
	if path, inMap := pathMap[reqPath]; inMap {
		return &path, nil
	}

	//get all keys with query params in path
	var pathUrls []string
	for k := range pathMap {
		if strings.Contains(k, "{") {
			pathUrls = append(pathUrls, k)
		}
	}

	//sort keys
	// TO-DO? Depends if we think that there are multiple routes that really match a given input

	//convert keys to regex
	pathsAsRegExp := make(map[string]string)
	re := regexp.MustCompile(`\{(.*?)\}`)
	for _, k := range pathUrls {
		pathsAsRegExp[k] = "^" + re.ReplaceAllLiteralString(k, `[^/]*`) + "$"
	}

	//iterate over and return first which matches
	for k, v := range pathsAsRegExp {
		if isMatch, err := regexp.MatchString(v, reqPath); err != nil {
			return nil, err
		} else if isMatch {
			if path, ok := pathMap[k]; ok {
				return &path, nil
			}
		}
	}
	return nil, errors.New("Path could not be matched to any in swagger spec.")
}

// Walks the whole json swagger schema and replaces any refs that just start with # with thef ull canonical ref
func ForceCanonicalRefs(swag *Schema, obj interface{}) {
	switch obj.(type) {
	case map[string]interface{}:
		// Replace any refs in the map
		m := obj.(map[string]interface{})
		if v, inMap := m["$ref"]; inMap && strings.HasPrefix(v.(string), "#") {
			m["$ref"] = "http://" + swag.Host + "/api-docs" + v.(string)
		}

		// Call this method with any other maps/arrays in the map.  Allof/AnyOf types register as slices here.
		for _, v := range m {
			if reflect.ValueOf(v).Kind() == reflect.Map || reflect.ValueOf(v).Kind() == reflect.Slice {
				ForceCanonicalRefs(swag, v)
			}
		}
	case []interface{}:
		for _, v := range obj.([]interface{}) {
			ForceCanonicalRefs(swag, v)
		}
	}
}

// Get the default swagger spec (URL tbd)
func GetSwaggerSpec() (*Schema, error) {
	return GetSwaggerSpecFromURL("http://mockapidemo.john.bagel.buffalo.im/api-docs")
}

func GetSwaggerSpecFromBytes(b []byte) (*Schema, error) {
	var swaggerSpec *Schema
	err := json.Unmarshal(b, &swaggerSpec)
	if err != nil {
		return nil, err
	}
	return swaggerSpec, nil
}

func GetSwaggerSpecFromURL(url string) (*Schema, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return GetSwaggerSpecFromBytes(body)
}

// TODO decide how to better pass in/handle host rather than hardcoding it (or update to location of newer specs).
// This project is getting dropped for now due to just validating remotely using a proxy server.
func ValidatePartialSchema(swag *Schema, endpointSchema map[string]interface{}, body string) (*gojsonschema.Result, error) {
	swaggerSpec := gojsonschema.NewReferenceLoader("http://" + swag.Host + "/api-docs")
	schemaToValidateAgainst := gojsonschema.NewGoLoader(endpointSchema)
	doc := gojsonschema.NewStringLoader(body)

	return gojsonschema.ValidatePartialSchema(swaggerSpec, schemaToValidateAgainst, doc)
}
