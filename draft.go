package gojsonschema

type Draft int

const (
	Hybrid Draft = 0
	Draft4 Draft = 4
	Draft6 Draft = 6
	Draft7 Draft = 7
)

var (
	draftDict = map[Draft]string{
		Draft4: "http://json-schema.org/draft-04/schema",
		Draft6: "http://json-schema.org/draft-06/schema",
		Draft7: "http://json-schema.org/draft-07/schema",
	}
	schemaDict = map[string]Draft{
		"http://json-schema.org/draft-04/schema": Draft4,
		"http://json-schema.org/draft-06/schema": Draft6,
		"http://json-schema.org/draft-07/schema": Draft7,
	}
)
