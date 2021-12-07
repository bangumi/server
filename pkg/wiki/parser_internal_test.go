package wiki

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseComponentSingleLine(t *testing.T) {
	t.Parallel()

	// assert.EqualValues(t, []string{"1 ", " 3 ", " 4"}, strings.SplitN("1 = 3 = 4", "=", 2))

	value, err := parseComponent(trim(`
	
	a = `))

	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, Field{Key: "a", Null: true}, *value)
}

func TestParseComponentMultiEq(t *testing.T) {
	t.Parallel()
	value, err := parseComponent(`a = b = c`)

	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t, Field{Key: "a", Value: "b = c"}, *value)
}

func TestParseComponentArray1(t *testing.T) {
	t.Parallel()
	value, err := parseComponent(trim(` 
a = {
[1]
[3]
}`))

	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t,
		Field{
			Key:   "a",
			Array: true,
			Values: []Item{
				{Value: "1"},
				{Value: "3"},
			},
		},
		*value)
}

func TestParseComponentArray2(t *testing.T) {
	t.Parallel()
	value, err := parseComponent(`a = {
[1]
[3]
}`)

	if !assert.Nil(t, err) {
		return
	}
	assert.Equal(t,
		Field{
			Key:   "a",
			Array: true,
			Values: []Item{
				{Value: "1"},
				{Value: "3"},
			},
		},
		*value)
}
