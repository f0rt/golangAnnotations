package repository

import (
	"fmt"
	"log"
	"strings"
	"text/template"
	"unicode"

	generator "github.com/MarcGrol/golangAnnotations/codegeneration"
	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/generationUtil"
	"github.com/MarcGrol/golangAnnotations/codegeneration/repository/repositoryAnnotation"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

type Generator struct {
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func (eg *Generator) GetAnnotations() []annotation.AnnotationDescriptor {
	return repositoryAnnotation.Get()
}

func (eg *Generator) Generate(inputDir string, parsedSource intermediatemodel.ParsedSources) error {
	structs := parsedSource.Structs

	packageName, err := generationUtil.GetPackageNameForStructs(structs)
	if packageName == "" || err != nil {
		return err
	}
	targetDir, err := generationUtil.DetermineTargetPath(inputDir, packageName)
	if err != nil {
		return err
	}
	for _, repository := range structs {
		if IsRepository(repository) {
			err = generationUtil.Generate(generationUtil.Info{
				Src:            fmt.Sprintf("%s.%s", repository.PackageName, repository.Name),
				TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/%s.go", targetDir, toFirstLower(repository.Name))),
				TemplateName:   "repository",
				TemplateString: repositoryTemplate,
				FuncMap:        customTemplateFuncs,
				Data:           repository,
			})
			if err != nil {
				log.Fatalf("Error generating repository %s: %s", repository.Name, err)
				return err
			}
		}
	}
	return nil
}

var customTemplateFuncs = template.FuncMap{
	"IsRepository":              IsRepository,
	"AggregateNameConst":        AggregateNameConst,
	"LowerAggregateName":        LowerAggregateName,
	"UpperAggregateName":        UpperAggregateName,
	"GetPackageName":            GetPackageName,
	"LowerModelName":            LowerModelName,
	"UpperModelName":            UpperModelName,
	"ModelPackageName":          ModelPackageName,
	"HasMethodFind":             HasMethodFind,
	"HasMethodFilterByEvent":    HasMethodFilterByEvent,
	"HasMethodFilterByMoment":   HasMethodFilterByMoment,
	"HasMethodFindStates":       HasMethodFindStates,
	"HasMethodExists":           HasMethodExists,
	"HasMethodAllAggregateUIDs": HasMethodAllAggregateUIDs,
	"HasMethodGetAllAggregates": HasMethodGetAllAggregates,
	"HasMethodPurgeOnEventUIDs": HasMethodPurgeOnEventUIDs,
	"HasMethodPurgeOnEventType": HasMethodPurgeOnEventType,
	"HasMethodPurgeAll":         HasMethodPurgeAll,
}

func IsRepository(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(repositoryAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(s.DocLines, repositoryAnnotation.TypeRepository)
	return ok
}

func AggregateNameConst(s intermediatemodel.Struct) string {
	return fmt.Sprintf("%sAggregateName", UpperAggregateName(s))
}

func LowerAggregateName(s intermediatemodel.Struct) string {
	return toFirstLower(GetAggregateName(s))
}

func UpperAggregateName(s intermediatemodel.Struct) string {
	return toFirstUpper(GetAggregateName(s))
}

func GetAggregateName(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(repositoryAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, repositoryAnnotation.TypeRepository); ok {
		return ann.Attributes[repositoryAnnotation.ParamAggregate]
	}
	return ""
}

func GetPackageName(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(repositoryAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, repositoryAnnotation.TypeRepository); ok {
		packageName := ann.Attributes[repositoryAnnotation.ParamPackage]
		if packageName != "" {
			return packageName
		}
	}
	return fmt.Sprintf("%sEvents", LowerAggregateName(s))
}

func LowerModelName(s intermediatemodel.Struct) string {
	return toFirstLower(GetModelName(s))
}

func UpperModelName(s intermediatemodel.Struct) string {
	return toFirstUpper(GetModelName(s))
}

func ModelPackageName(s intermediatemodel.Struct) string {
	return toFirstLower(GetModelName(s)) + "Model"
}

func GetModelName(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(repositoryAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, repositoryAnnotation.TypeRepository); ok {
		m := ann.Attributes[repositoryAnnotation.ParamModel]
		if m != "" {
			return m
		}
	}
	return GetAggregateName(s)
}

func HasMethodFind(s intermediatemodel.Struct) bool {
	return HasMethod(s, "find")
}

func HasMethodFilterByEvent(s intermediatemodel.Struct) bool {
	return HasMethod(s, "filterByEvent")
}

func HasMethodFilterByMoment(s intermediatemodel.Struct) bool {
	return HasMethod(s, "filterByMoment")
}

func HasMethodFindStates(s intermediatemodel.Struct) bool {
	return HasMethod(s, "findStates")
}

func HasMethodExists(s intermediatemodel.Struct) bool {
	return HasMethod(s, "exists")
}

func HasMethodAllAggregateUIDs(s intermediatemodel.Struct) bool {
	return HasMethod(s, "allAggregateUIDs")
}

func HasMethodGetAllAggregates(s intermediatemodel.Struct) bool {
	return HasMethod(s, "allAggregates")
}

func HasMethodPurgeOnEventUIDs(s intermediatemodel.Struct) bool {
	return HasMethod(s, "purgeOnEventUIDs")
}

func HasMethodPurgeOnEventType(s intermediatemodel.Struct) bool {
	return HasMethod(s, "purgeOnEventType")
}

func HasMethodPurgeAll(s intermediatemodel.Struct) bool {
	return HasMethod(s, "purgeAll")
}

func HasMethod(s intermediatemodel.Struct, methodName string) bool {
	annotations := annotation.NewRegistry(repositoryAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, repositoryAnnotation.TypeRepository); ok {
		methods := strings.Split(ann.Attributes[repositoryAnnotation.ParamMethods], ",")
		for _, method := range methods {
			if strings.TrimSpace(method) == methodName {
				return true
			}
		}
	}
	return false
}

func toFirstLower(in string) string {
	a := []rune(in)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func toFirstUpper(in string) string {
	a := []rune(in)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}
