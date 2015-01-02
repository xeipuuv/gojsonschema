[![Build Status](https://travis-ci.org/xeipuuv/gojsonschema.svg)](https://travis-ci.org/xeipuuv/gojsonschema)

# gojsonschema

## Description

An implementation of JSON Schema, based on IETF's draft v4 - Go language

References :

* http://json-schema.org
* http://json-schema.org/latest/json-schema-core.html
* http://json-schema.org/latest/json-schema-validation.html

## Installation

```
go get github.com/xeipuuv/gojsonschema
```

Dependencies :
* https://github.com/xeipuuv/gojsonpointer
* https://github.com/xeipuuv/gojsonreference
* https://github.com/stretchr/testify/assert

## Usage 

### Example

```go

package main

import (
    "fmt"
    gjs "github.com/xeipuuv/gojsonschema"
)

func main() {

    schema, err := gjs.NewSchema("file:///home/me/schema.json")
    if err != nil {
        panic(err.Error())
    }

    document, err := gjs.GetFile("/home/me/document.json")
    if err != nil {
        panic(err.Error())
    }

    result, err := schema.Validate(document)
    if err != nil {
        panic(err.Error())
    }

    if result.Valid() {
        fmt.Printf("The document is valid\n")
    } else {
        fmt.Printf("The document is not valid. see errors :\n")
        for _, desc := range result.Errors() {
            fmt.Printf("- %s\n", desc)
        }
    }

}


```

#### Loading a schema

* Using HTTP :

```go
    schema, err := gjs.NewSchema("http://www.some_host.com/schema.json")
```

* Using a local file :

```go
    schema, err := gjs.NewSchema("file:///home/some_user/schema.json")
```

Note the URI scheme is used here (file://), also the full path to the file is required.

* Using a Go map[string]interface{} :

```go
    m := map[string]interface{}{"type": "string"}
    schema, err := gjs.NewSchema(m)
```

* Using a JSON string :

```go
    schema, err := gjs.NewSchema(`{"type": "string"}`)
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
	document := `{"name": "john"}`
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


## Uses

gojsonschema uses the following test suite :

https://github.com/json-schema/JSON-Schema-Test-Suite
