package rest

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
	os.Remove(generationUtil.Prefixed("./testData/httpMyService.go"))
	os.Remove(generationUtil.Prefixed("./testData/httpMyServiceHelpers_test.go"))
	os.Remove(generationUtil.Prefixed("./testData/httpClientForMyService.go"))
	os.Remove(generationUtil.Prefixed("./testData/httpMyServiceHelpers_test.go"))
	os.Remove(generationUtil.Prefixed("./testData/testDataTestLog/httpTestMyService.go"))
}

func TestGenerateForWeb(t *testing.T) {
	cleanup()
	defer cleanup()

	s := []intermediatemodel.Struct{
		{
			DocLines:    []string{"// @RestService( path = \"/api\")"},
			PackageName: "testData",
			Name:        "MyService",
			Operations:  []*intermediatemodel.Operation{},
		},
	}

	s[0].Operations = append(s[0].Operations,
		&intermediatemodel.Operation{
			DocLines:      []string{"// @RestOperation(path = \"/person\", method = \"GET\", format = \"JSON\", form = \"true\" )"},
			Name:          "doit",
			RelatedStruct: &intermediatemodel.Field{TypeName: "MyService"},
			InputArgs: []intermediatemodel.Field{
				{Name: "uid", TypeName: "int"},
				{Name: "subuid", TypeName: "string"},
			},
			OutputArgs: []intermediatemodel.Field{
				{TypeName: "error"},
			},
		})
	{
		err := NewGenerator().Generate("testData", intermediatemodel.ParsedSources{Structs: s})
		assert.Nil(t, err)
	}

	{
		{
			// check that generated files exists
			_, err := os.Stat(generationUtil.Prefixed("./testData/httpMyService.go"))
			assert.NoError(t, err)
		}
		{
			// check that generate code has 4 helper functions for MyStruct
			data, err := ioutil.ReadFile(generationUtil.Prefixed("./testData/httpMyService.go"))
			assert.NoError(t, err)
			assert.Contains(t, string(data), "func (ts *MyService) HTTPHandler() http.Handler {")
			assert.Contains(t, string(data), "func doit(service *MyService) http.HandlerFunc {")
		}
	}
	{
		{
			// check that generated files exists
			_, err := os.Stat(generationUtil.Prefixed("./testData/httpMyService.go"))
			assert.NoError(t, err)
		}
		{
			// check that generate code has 4 helper functions for MyStruct
			data, err := ioutil.ReadFile(generationUtil.Prefixed("./testData/httpMyServiceHelpers_test.go"))
			assert.NoError(t, err)
			assert.Contains(t, string(data), "func doitTestHelper")
		}
	}

}
