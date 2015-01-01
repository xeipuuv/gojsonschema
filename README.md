[![Build Status](https://travis-ci.org/jabley/gojsonschema.svg?branch=master)](https://travis-ci.org/jabley/gojsonschema)

# gojsonschema

## Description
An implementation of JSON Schema, based on IETF's draft v4 - Go language

## Usage 

### Quick example

```go

package main

import (
    "fmt"
    gjs "github.com/xeipuuv/gojsonschema"
)

func main() {

    // Loads a schema remotely
    schema, err := gjs.NewSchema("http://host/schema.json")
    if err != nil {
        panic(err.Error()) // could not read from HTTP for example
    }

    // Loads the JSON to validate from a local file
    document, err := gjs.GetFile("/home/me/data.json")
    if err != nil {
        panic(err.Error()) // could be file not found on your hard drive
    }

	// Try to validate the JSON against the schema
    result, err := schema.Validate(document)
    if err != nil {
        panic(err.Error()) // the document is not valid JSON
    }

	// Deal with result
    if result.Valid() {
    
        fmt.Printf("The document is valid\n")
    
    } else {
    
        fmt.Printf("The document is not valid. see errors :\n")
    
        // Loop through errors
        for _, desc := range result.Errors() {
            fmt.Printf("- %s\n", desc)
        }
    
    }

}


```

#### Loading a schema

Schemas can be loaded remotely from a HTTP Url:

```go
    schemaDocument, err := gjs.NewSchema("http://myhost/schema.json")
```

From a local file, using the file URI scheme:

```go
	schemaDocument, err := gjs.NewSchema("file:///home/me/schema.json")
```


You may also load the schema from within your code, using a map[string]interface{} variable or a JSON string.

Note that schemas loaded from non-HTTP are subject to limitations, they need to be standalone schemas; 
Which means references to local files and/or remote files within these schemas will not work.

```go
	schemaMap := map[string]interface{}{
		"type": "string"}

	schema, err := gojsonschema.NewSchema(schemaMap)

	// or using a string
	// schema, err := gojsonschema.NewSchema("{\"type\": \"string\"}")

```

#### Loading a JSON

The library virtually accepts any JSON since it uses reflection to validate against the schema.

You may use and combine go types like 
* string (JSON string)
* bool (JSON boolean)
* float64 (JSON number)
* nil (JSON null)
* slice (JSON array)
* map[string]interface{} (JSON object)

You can declare your Json from within your code, using a map / interface{}:

```go
	jsonDocument := map[string]interface{}{
		"name": "john"}
```

A string:

```go
	jsonDocument := "{\"name\": \"john\"}"
```

Helper functions are also available to load from a Http URL:

```go
    jsonDocument, err := gojsonschema.GetHTTP("http://host/data.json")
```

Or a local file:

```go
	jsonDocument, err := gojsonschema.GetFile("/home/me/data.json")
```

#### Validation

Once the schema and the JSON to validate are loaded, validation phase becomes easy:

```go
	result, err := schemaDocument.Validate(jsonDocument)
```

Check the result validity with:

```go
	if result.Valid() {
		// Your Json is valid
	}
```

If not valid, you can loop through the error messages returned by the validation phase:

```go
	for _, desc := range result.Errors() {
    	fmt.Printf("Error: %s\n", desc)
	}
```

## Dependencies
https://github.com/xeipuuv/gojsonpointer

https://github.com/xeipuuv/gojsonreference

https://github.com/stretchr/testify/assert

## Uses

gojsonschema uses the following test suite :

https://github.com/json-schema/JSON-Schema-Test-Suite

## References

###Website
http://json-schema.org

###Schema Core
http://json-schema.org/latest/json-schema-core.html

###Schema Validation
http://json-schema.org/latest/json-schema-validation.html
