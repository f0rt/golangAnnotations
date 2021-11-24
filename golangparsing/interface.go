package golangparsing

import "github.com/MarcGrol/golangAnnotations/intermediatemodel"

type Parser interface {
	ParseSourceDir(dirName string, includeRegex string, excludeRegex string) (intermediatemodel.ParsedSources, error)
}
