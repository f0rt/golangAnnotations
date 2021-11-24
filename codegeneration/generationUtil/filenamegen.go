package generationUtil

import (
	"path"

	generator "github.com/MarcGrol/golangAnnotations/codegeneration"
)

func Prefixed(filenamePath string) string {
	dir, filename := path.Split(filenamePath)
	return dir + generator.GenfilePrefix + filename
}
