package gojsonschema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testMessageMaker = newMessageMaker("BASIC_ERROR", "%v %v")
)

func TestMessageMakerFormat(t *testing.T) {
	message := testMessageMaker("one", "two")
	assert.True(t, message.Description == "one two")
}

func TestMessageMakerError(t *testing.T) {
	message := testMessageMaker("one", "two")
	assert.True(t, message.Error() != nil)
	assert.True(t, message.Error().Error() == "one two")
}
