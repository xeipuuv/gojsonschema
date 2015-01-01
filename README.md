[![Build Status](https://travis-ci.org/jabley/gojsonschema.svg?branch=master)](https://travis-ci.org/jabley/gojsonschema)

# gojsonschema

## Description
An implementation of JSON Schema, based on IETF's draft v4 - Go language

## Installation

```
go get github.com/xeipuuv/gojsonschema
```

## Usage 

### Quick example

```go

package main

import (
    "fmt"
    gjs "github.com/xeipuuv/gojsonschema"
)

func main() {

    // loads a schema from the Web
    schema, err := gjs.NewSchema("http://host/schema.json")
    if err != nil {
        panic(err.Error()) // could be : invalid address or timeout
    }

    // loads the JSON to validate from a local file
    document, err := gjs.GetFile("/home/me/data.json")
    if err != nil {
        panic(err.Error()) // could be : file not found on your hard drive
    }

	// try to validate the JSON against the schema
    result, err := schema.Validate(document)
    if err != nil {
        panic(err.Error()) // could be : the document is not valid JSON
    }

	// deal with result
    if result.Valid() {
    
        fmt.Printf("The document is valid\n")
    
    } else {
    
        fmt.Printf("The document is not valid. see errors :\n")
    
        // display validation errors
        for _, desc := range result.Errors() {
            fmt.Printf("- %s\n", desc)
        }
    
    }

}


```

#### Loading a schema

Schemas can be loaded remotely from a HTTP URL :

```go
    schema, err := gjs.NewSchema("http://myhost/schema.json")
```

From a local file, using the file URI scheme:

```go
	schema, err := gjs.NewSchema("file:///home/me/schema.json")
```


You may also load the schema from within your code, using a map[string]interface{} variable or a JSON string.

Note that schemas loaded from non-HTTP are subject to limitations, they need to be standalone schemas; 
That means references to local files and/or remote files within these schemas will not work.

```go
	m := map[string]interface{}{
		"type": "string"}

	schema, err := gjs.NewSchema(m)

	// or using a string
	// schema, err := gjs.NewSchema(`{"type": "string"}`)
```

#### Loading a JSON

The library virtually accepts any form of JSON since it uses reflection to validate against the schema.

You may use and combine go types like :

* string (JSON string)
* bool (JSON boolean)
* float64 (JSON number)
* nil (JSON null)
* slice (JSON array)
* map[string]interface{} (JSON object)

You can declare your JSON from within your code, using a map / interface{} :

```go
	document := map[string]interface{}{
		"name": "john"}
```

Or a JSON string:

```go
	document := "{\"name\": \"john\"}"
```

Helper functions are also available to load from a HTTP URL :

```go
    document, err := gjs.GetHTTP("http://host/data.json")
```

Or a local file :

```go
	document, err := gjs.GetFile("/home/me/data.json")
```

#### Validation

Once the schema and the JSON to validate are loaded, validation phase becomes easy :

```go
	result, err := schema.Validate(document)
```

Check the result with:

```go
	if result.Valid() {
		// Your Json is valid
	}
```

If not valid, you can loop through the error messages returned by the validation phase :

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
