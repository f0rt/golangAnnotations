package jsonHelpers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/MarcGrol/golangAnnotations/codegeneration/generationUtil"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
	"github.com/stretchr/testify/assert"
)

func cleanup() {
	os.Remove(generationUtil.Prefixed("./testData/ast.json"))
	os.Remove(generationUtil.Prefixed("./testData/example_json.go"))
}

func TestGenerateForJson(t *testing.T) {
	cleanup()
	defer cleanup()

	e := []intermediatemodel.Enum{
		{
			PackageName: "testData",
			Filename:    "example.go",
			DocLines:    []string{"// @JsonEnum()"},
			Name:        "ColorType",
			EnumLiterals: []intermediatemodel.EnumLiteral{
				{Name: "ColorTypeRed"},
				{Name: "ColorTypeGreen"},
				{Name: "ColorTypeBlue"},
			},
		},
	}

	s := []intermediatemodel.Struct{
		{
			PackageName: "testData",
			Filename:    "example.go",
			DocLines:    []string{`// @JsonStruct()`},
			Name:        "ColoredThing",
			Fields: []intermediatemodel.Field{
				{
					Name:     "Name",
					TypeName: "string",
				},
				{
					Name:     "Tags",
					TypeName: "[]string",
				},
				{
					Name:     "PrimaryColor",
					TypeName: "ColorType",
				},
				{
					Name:     "OtherColors",
					TypeName: "[]ColorType",
				},
			},
		},
	}

	ps := intermediatemodel.ParsedSources{
		Enums:   e,
		Structs: s,
	}
	err := NewGenerator().Generate("./testData/", ps)
	assert.Nil(t, err)

	// check that generated files exists
	_, err = os.Stat(generationUtil.Prefixed("./testData/example_json.go"))
	assert.NoError(t, err)

	// check that generate code has 4 helper functions for MyStruct
	data, err := ioutil.ReadFile(generationUtil.Prefixed("./testData/example_json.go"))
	assert.NoError(t, err)
	assert.Contains(t, string(data), `func (t *ColorType) UnmarshalJSON(data []byte) error {`)
	assert.Contains(t, string(data), `func (t ColorType) MarshalJSON() ([]byte, error) {`)

	assert.Contains(t, string(data), `func (data *ColoredThing) UnmarshalJSON(b []byte) error {`)
	assert.Contains(t, string(data), `func (data ColoredThing) MarshalJSON() ([]byte, error) {`)

}

func TestIsJsonEnum(t *testing.T) {
	e := intermediatemodel.Enum{
		DocLines: []string{
			`// @JsonStruct()`,
			`// @JsonEnum()`,
		},
	}
	assert.True(t, IsJSONEnum(e))
}

func TestIsJsonStruct(t *testing.T) {
	s := intermediatemodel.Struct{
		DocLines: []string{
			`// @Event(aggregate = "Test")`,
			`// @JsonStruct()`,
		},
	}
	assert.True(t, IsJSONStruct(s))
}
