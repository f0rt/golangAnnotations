package generator

import (
	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

const (
	GenfilePrefix       = "gen_"
	GenfileExcludeRegex = GenfilePrefix + ".*"
)

type Generator interface {
	GetAnnotations() []annotation.AnnotationDescriptor
	Generate(inputDir string, parsedSources intermediatemodel.ParsedSources) error
}
