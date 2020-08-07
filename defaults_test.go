package gojsonschema

import (
	"log"
	"testing"
)

func M(in ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if len(in)%2 != 0 {
		log.Fatal("map construction M must have one value for each key")
	}

	for i := 0; i < len(in); i += 2 {
		k := in[i]
		v := in[i+1]
		sK := k.(string)
		result[sK] = v
	}

	return result
}

// objSchemaFromProperties constructs a schema from "properties"
func objSchemaFromProperties(properties map[string]interface{}) *Schema {
	schemaMap := M("type", "object", "properties", properties)
	loader := NewGoLoader(schemaMap)

	// Since its a test, it'll just fail and then we address later
	schema, _ := NewSchema(loader)

	return schema
}

// arrSchemaFromProperties makes sure that each item in an array contains
// the properties passed in "properties"
func arrSchemaFromProperties(properties map[string]interface{}) *Schema {
	objMap := M("type", "object", "properties", properties)
	arrMap := M("type", "array", "items", objMap)

	loader := NewGoLoader(arrMap)

	// Since its a test, it'll just fail and then we address later
	schema, _ := NewSchema(loader)

	return schema
}

// InsertDefaults fails when nil is passed as arguments
func TestInsertNil(t *testing.T) {
	properties := M()
	schema := objSchemaFromProperties(properties)

	_, err := schema.InsertDefaults(nil)
	if err == nil {
		t.Error("InsertDefault should fail with a nil argument")
	}
}

// InsertDefaults succeeds when empty map is passed as argument
func TestSimpleDefault(t *testing.T) {
	properties := M("testkey", M("default", "defaultvalue"))
	schema := objSchemaFromProperties(properties)

	into := make(map[string]interface{})

	r, err := schema.InsertDefaults(into)
	if err != nil {
		t.Error(err)
	}

	result := r.(map[string]interface{})

	if v := result["testkey"]; v != "defaultvalue" {
		t.Error("InsertDefaults failed to add 'defaultvalue' at 'testkey'")
	}
}

func TestDoesNotOverwrite(t *testing.T) {
	properties := M("testkey", M("default", "defaultvalue"))
	schema := objSchemaFromProperties(properties)

	into := make(map[string]interface{})
	into["testkey"] = "someothervalue"

	r, err := schema.InsertDefaults(into)
	if err != nil {
		t.Error(err)
	}

	result := r.(map[string]interface{})

	if v := result["testkey"]; v != "someothervalue" {
		t.Error("InsertDefaults has overwritten a value that was there before")
	}
}

// TestNestedValues makes sure that a default value several layers deep will be inserted
func TestNestedValues(t *testing.T) {
	properties := M("deep", M("type", "object", "properties", M("testkey", M("default", "defaultvalue"))))
	schema := objSchemaFromProperties(properties)

	into := make(map[string]interface{})

	r, err := schema.InsertDefaults(into)
	if err != nil {
		t.Error(err)
	}

	result := r.(map[string]interface{})

	innerMap := result["deep"].(map[string]interface{})

	if v := innerMap["testkey"]; v != "defaultvalue" {
		t.Error("InsertDefaults failed to add 'defaultvalue' at .'deep'.'testkey'")
	}
}

// If an empty array is passed, nothing should be inserted, even if there
// is a default value specified somewhere
func TestSimpleArr(t *testing.T) {
	properties := M("testkey", M("default", "defaultvalue"))
	schema := arrSchemaFromProperties(properties)

	examplearr := make([]map[string]interface{}, 0)

	_, err := schema.InsertDefaults(examplearr)
	if err != nil {
		t.Error(err)
	}
}

func TestArrayOfProperties(t *testing.T) {
	properties := M("testkey", M("default", "defaultvalue"))
	schema := arrSchemaFromProperties(properties)

	emptyexample := M()

	examplearr := make([]map[string]interface{}, 1)
	examplearr[0] = emptyexample

	r, err := schema.InsertDefaults(examplearr)
	if err != nil {
		t.Error(err)
	}

	if res := r.([]map[string]interface{}); res[0]["testkey"] != "defaultvalue" {
		t.Error("array[0] was not filled with the proper default values")
	}
}
