# gojsonschema

## Description
An implementation of JSON Schema, based on IETF's draft v4 - Go language

## Status

Functional, one feature is missing : id(s) as scope for references

Test phase : Passed 99.19% of Json Schema Test Suite

## Usage 

### Basic example

```

package main

import (
    "fmt"
    "github.com/sigu-399/gojsonschema"
)

func main() {

    // use a remote schema
    schema, err := gojsonschema.NewJsonSchemaDocument("http://myhost/schema1.json")
    // ... or a local file
    //schema, err := gojsonschema.NewJsonSchemaDocument("file:///home/me/myschemas/schema1.json")
    if err != nil {
        panic(err.Error())
    }

    // use a remote json to validate
    jsonToValidate, err := gojsonschema.GetHttpJson("http://myhost/someDoc1.json")
    // ... or a local one
    //jsonToValidate, err := gojsonschema.GetFileJson("/home/me/mydata/someDoc1.json")

    if err != nil {
        panic(err.Error())
    }

    validationResult := schema.Validate(jsonToValidate)

    if validationResult.IsValid() {

        fmt.Printf("The document is valid\n")

    } else {

        fmt.Printf("The document is not valid. see errors :\n")
        for _, errorMessage := range validationResult.GetErrorMessages() {
            fmt.Printf("- %s\n", errorMessage)
        }

    }

}


```

## References

###Website
http://json-schema.org

###Schema Core
http://json-schema.org/latest/json-schema-core.html

###Schema Validation
http://json-schema.org/latest/json-schema-validation.html

## Dependencies
https://github.com/sigu-399/gojsonpointer

https://github.com/sigu-399/gojsonreference

## Uses

gojsonschema uses the following test suite :

https://github.com/json-schema/JSON-Schema-Test-Suite
