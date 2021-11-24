package rest

import (
	"fmt"
	"testing"

	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
	"github.com/stretchr/testify/assert"
)

func TestIsRestService(t *testing.T) {
	s := intermediatemodel.Struct{
		DocLines: []string{
			`//@RestService( path = "/api")`},
	}
	assert.True(t, IsRestService(s))
}

func TestGetRestServicePath(t *testing.T) {
	s := intermediatemodel.Struct{
		DocLines: []string{
			`//@RestService( path = "/api")`},
	}
	assert.Equal(t, "/api", GetRestServicePath(s))
}

func TestIsRestOperation(t *testing.T) {
	assert.True(t, IsRestOperation(createOper("GET")))
}

func TestGetRestOperationMethod(t *testing.T) {
	assert.Equal(t, "GET", GetRestOperationMethod(createOper("GET")))
}

func TestGetRestOperationPath(t *testing.T) {
	assert.Equal(t, "/api/person", GetRestOperationPath(createOper("DONTCARE")))
}

func TestHasInputGet(t *testing.T) {
	assert.False(t, HasInput(createOper("GET")))
}

func TestHasInputDelete(t *testing.T) {
	assert.False(t, HasInput(createOper("DELETE")))
}

func TestHasInputPost(t *testing.T) {
	assert.True(t, HasInput(createOper("POST")))
}

func TestHasInputPut(t *testing.T) {
	assert.True(t, HasInput(createOper("PUT")))
}

func TestGetInputArgTypeString(t *testing.T) {
	o := intermediatemodel.Operation{
		InputArgs: []intermediatemodel.Field{
			{TypeName: "string"},
		},
	}
	assert.Equal(t, "", GetInputArgType(o))
}

func TestGetInputArgTypePerson(t *testing.T) {
	assert.Equal(t, "Person", GetInputArgType(createOper("DONTCARE")))
}

func TestGetInputArgName(t *testing.T) {
	assert.Equal(t, "person", GetInputArgName(createOper("DONTCARE")))
}

func TestGetInputParamString(t *testing.T) {
	assert.Equal(t, "ctx, uid, person", GetInputParamString(createOper("DONTCARE")))
}

func TestHasOutput(t *testing.T) {
	assert.True(t, HasOutput(createOper("DONTCARE")))
}

func TestGetOutputArgType(t *testing.T) {
	assert.Equal(t, "Person", GetOutputArgType(createOper("DONTCARE")))
}

func TestIsPrimitiveTrue(t *testing.T) {
	f := intermediatemodel.Field{Name: "uid", TypeName: "string"}
	assert.False(t, IsCustomArg(f))
}

func TestIsPrimitiveFalse(t *testing.T) {
	f := intermediatemodel.Field{Name: "person", TypeName: "Person"}
	assert.True(t, IsCustomArg(f))
}

func TestIsNumberTrue(t *testing.T) {
	f := intermediatemodel.Field{Name: "uid", TypeName: "int"}
	assert.True(t, IsIntArg(f))
}

func TestIsNumberFalse(t *testing.T) {
	f := intermediatemodel.Field{Name: "uid", TypeName: "string"}
	assert.False(t, IsIntArg(f))
}

func createOper(method string) intermediatemodel.Operation {
	o := intermediatemodel.Operation{
		DocLines: []string{
			fmt.Sprintf("//@RestOperation( method = \"%s\", path = \"/api/person\")", method),
		},
		InputArgs: []intermediatemodel.Field{
			{Name: "ctx", TypeName: "context.Context"},
			{Name: "uid", TypeName: "string"},
			{Name: "person", TypeName: "Person"},
		},
		OutputArgs: []intermediatemodel.Field{
			{TypeName: "Person"},
			{TypeName: "error"},
		},
	}
	return o
}
