[![Build Status](https://travis-ci.org/xeipuuv/gojsonschema.svg)](https://travis-ci.org/xeipuuv/gojsonschema)

# gojsonschema

## Description

An implementation of JSON Schema, based on IETF's draft v4 - Go language

## NOTE:

This repo was forked to support validating against partial schemas and fragments to facilitate
client-side validation of Apcera's goswagger schema.  Our current approach, however, diverted
to using a proxy server and we no longer required this code to execute the testing locally.

This code is still being merged into our local repo and a PR was submitted to the upstream
(https://github.com/xeipuuv/gojsonschema/pull/55).

References :

* http://json-schema.org
* http://json-schema.org/latest/json-schema-core.html
* http://json-schema.org/latest/json-schema-validation.html

## Installation

```
go get github.com/xeipuuv/gojsonschema
```

Dependencies :
* [github.com/xeipuuv/gojsonpointer](https://github.com/xeipuuv/gojsonpointer)
* [github.com/xeipuuv/gojsonreference](https://github.com/xeipuuv/gojsonreference)
* [github.com/stretchr/testify/assert](https://github.com/stretchr/testify#assert-package)

## Usage 

### Example

```go

package main

import (
    "fmt"
    "github.com/xeipuuv/gojsonschema"
)

func main() {

    schemaLoader := gojsonschema.NewReferenceLoader("file:///home/me/schema.json")
    documentLoader := gojsonschema.NewReferenceLoader("file:///home/me/document.json")

    result, err := gojsonschema.Validate(schemaLoader, documentLoader)
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

#### Loaders

There are various ways to load your JSON data.
In order to load your schemas and documents, 
first declare an appropriate loader :

* Web / HTTP, using a reference :

```go
loader := gojsonschema.NewReferenceLoader("http://www.some_host.com/schema.json")
```

* Local file, using a reference :

```go
loader := gojsonschema.NewReferenceLoader("file:///home/me/schema.json")
```

References use the URI scheme, the prefix (file://) and a full path to the file are required.

* JSON strings :

```go
loader := gojsonschema.NewStringLoader(`{"type": "string"}`)
```

* Custom Go types :

```go
m := map[string]interface{}{"type": "string"}
loader := gojsonschema.NewGoLoader(m)
```

And

```go
type Root struct {
	Users []User `json:"users"`
}

type User struct {
	Name string `json:"name"`
}

...

data := Root{}
data.Users = append(data.Users, User{"John"})
data.Users = append(data.Users, User{"Sophia"})
data.Users = append(data.Users, User{"Bill"})

loader := gojsonschema.NewGoLoader(data)
```

#### Validation

Once the loaders are set, validation is easy :

```go
result, err := gojsonschema.Validate(schemaLoader, documentLoader)
```

Alternatively, you might want to load a schema only once and process to multiple validations :

```go
schema, err := gojsonschema.NewSchema(schemaLoader)
...
result1, err := schema.Validate(documentLoader1)
...
result2, err := schema.Validate(documentLoader2)
...
// etc ...
```

To check the result :

```go
    if result.Valid() {
    	fmt.Printf("The document is valid\n")
    } else {
        fmt.Printf("The document is not valid. see errors :\n")
        for _, desc := range result.Errors() {
            fmt.Printf("- %s\n", desc)
        }
    }
```

Validation with partial schemas is similar.  It allows the definitions and references in a main
schema to be used to resolve references and fragments in another document.  Currently, you must
manually make the references canonical. :

```
    mainSchema := gojsonschema.NewReferenceLoader(schemaURL)
    schemaFragment := gojsonschema.NewGoLoader(endpointSchema)
    doc := gojsonschema.NewStringLoader(someString)

    return ValidatePartialSchema(mainSchema, schemaFragment, doc)
```

## Uses

gojsonschema uses the following test suite :

https://github.com/json-schema/JSON-Schema-Test-Suite
