package jsonHelpers

import (
	"fmt"
	"log"
	"strings"
	"text/template"
	"unicode"

	generator "github.com/MarcGrol/golangAnnotations/codegeneration"
	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/generationUtil"
	"github.com/MarcGrol/golangAnnotations/codegeneration/jsonHelpers/jsonAnnotation"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

type Generator struct {
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func (eg *Generator) GetAnnotations() []annotation.AnnotationDescriptor {
	return jsonAnnotation.Get()
}

type jsonContext struct {
	PackageName string
	Enums       []intermediatemodel.Enum
	Structs     []intermediatemodel.Struct
}

func (eg *Generator) Generate(inputDir string, parsedSource intermediatemodel.ParsedSources) error {
	enums := parsedSource.Enums
	structs := parsedSource.Structs

	packageName, err := generationUtil.GetPackageNameForEnumsOrStructs(enums, structs)
	if packageName == "" || err != nil {
		return err
	}
	targetDir, err := generationUtil.DetermineTargetPath(inputDir, packageName)
	if err != nil {
		return err
	}

	jsonEnums := make([]intermediatemodel.Enum, 0, len(enums))
	for _, anEnum := range enums {
		if IsJSONEnum(anEnum) {
			jsonEnums = append(jsonEnums, anEnum)
		}
	}
	jsonStructs := make([]intermediatemodel.Struct, 0, len(structs))
	for _, aStruct := range structs {
		if IsJSONStruct(aStruct) {
			jsonStructs = append(jsonStructs, aStruct)
		}
	}
	if len(jsonEnums) == 0 && len(jsonStructs) == 0 {
		return nil
	}

	err = doGenerate(packageName, jsonEnums, jsonStructs, targetDir)
	if err != nil {
		return err
	}

	return nil
}

func doGenerate(packageName string, jsonEnums []intermediatemodel.Enum, jsonStructs []intermediatemodel.Struct, targetDir string) error {
	filenameMap := getFilenamesWithTypeNames(jsonEnums, jsonStructs)

	for fn := range filenameMap {
		targetFilename := strings.Replace(fn, ".", "_json.", 1)
		target := generationUtil.Prefixed(fmt.Sprintf("%s/%s", targetDir, targetFilename))

		data := jsonContext{
			PackageName: packageName,
		}

		// find al enums belonging to this file
		for _, e := range jsonEnums {
			if e.Filename == fn {
				data.Enums = append(data.Enums, e)
			}
		}
		for _, s := range jsonStructs {
			if s.Filename == fn {
				data.Structs = append(data.Structs, s)
			}
		}

		if len(data.Enums) > 0 || len(data.Structs) > 0 {
			err := generationUtil.Generate(generationUtil.Info{
				Src:            packageName,
				TargetFilename: target,
				TemplateName:   "json-enums",
				TemplateString: jsonHelpersTemplate,
				FuncMap:        customTemplateFuncs,
				Data:           data,
			})
			if err != nil {
				log.Fatalf("Error generating wrappers for enums (%s)", err)
				return err
			}
		}
	}

	return nil
}

func getFilenamesWithTypeNames(jsonEnums []intermediatemodel.Enum, jsonStructs []intermediatemodel.Struct) map[string][]string {
	// group enum and structs by filename
	filenameMap := map[string][]string{}

	// get all  enum-names belonging to file
	for _, e := range jsonEnums {
		typeNames := filenameMap[e.Filename]
		typeNames = append(typeNames, e.Name)
		filenameMap[e.Filename] = typeNames
	}

	// get all  struct-names belonging to file
	for _, s := range jsonStructs {
		typeNames := filenameMap[s.Filename]
		typeNames = append(typeNames, s.Name)
		filenameMap[s.Filename] = typeNames
	}

	return filenameMap
}

var customTemplateFuncs = template.FuncMap{
	"HasAlternativeName": hasAlternativeName,
	"GetAlternativeName": getAlternativeName,
	"GetPreferredName":   getPreferredName,
	"HasDefaultValue":    hasDefaultValue,
	"GetDefaultValue":    getDefaultValue,
	"HasSlices":          hasSlices,
}

func IsJSONEnum(e intermediatemodel.Enum) bool {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum)
	return ok
}

func IsJSONEnumStripped(e intermediatemodel.Enum) bool {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum); ok {
		return ann.Attributes[jsonAnnotation.ParamStripped] == "true"
	}
	return false
}

func IsJSONEnumLiteral(e intermediatemodel.Enum) bool {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum); ok {
		return ann.Attributes[jsonAnnotation.ParamLiteral] == "true"
	}
	return false
}

func IsJSONEnumTolerant(e intermediatemodel.Enum) bool {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum); ok {
		return ann.Attributes[jsonAnnotation.ParamTolerant] == "true"
	}
	return false
}

func GetJSONEnumBase(e intermediatemodel.Enum) string {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum); ok {
		return ann.Attributes[jsonAnnotation.ParamBase]
	}
	return ""
}

func HasJSONEnumBase(e intermediatemodel.Enum) bool {
	return GetJSONEnumBase(e) != ""
}

func GetJSONEnumDefault(e intermediatemodel.Enum) string {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(e.DocLines, jsonAnnotation.TypeEnum); ok {
		return ann.Attributes[jsonAnnotation.ParamDefault]
	}
	return ""
}

func hasDefaultValue(e intermediatemodel.Enum) bool {
	return GetJSONEnumDefault(e) != ""
}

func getDefaultValue(e intermediatemodel.Enum) string {
	return GetJSONEnumBase(e) + GetJSONEnumDefault(e)
}

func hasAlternativeName(e intermediatemodel.Enum) bool {
	return HasJSONEnumBase(e) && IsJSONEnumTolerant(e)
}

// special feature to work around literal names that should contain '-': use 'ɂ' instead
func fixedLitName(lit intermediatemodel.EnumLiteral) string {
	return strings.Replace(lit.Name, "ɂ", "-", -1)
}

func getAlternativeName(e intermediatemodel.Enum, lit intermediatemodel.EnumLiteral) string {
	name := fixedLitName(lit)
	if IsJSONEnumStripped(e) {
		return lowerInitialIfNeeded(e, name)
	}
	base := GetJSONEnumBase(e)
	return lowerInitialIfNeeded(e, strings.TrimPrefix(name, base))
}

func getPreferredName(e intermediatemodel.Enum, lit intermediatemodel.EnumLiteral) string {
	name := fixedLitName(lit)
	if IsJSONEnumStripped(e) {
		base := GetJSONEnumBase(e)
		return lowerInitialIfNeeded(e, strings.TrimPrefix(name, base))
	}
	return lowerInitialIfNeeded(e, name)
}

func lowerInitialIfNeeded(e intermediatemodel.Enum, s string) string {
	if IsJSONEnumLiteral(e) {
		return s
	}
	a := []rune(s)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func IsJSONStruct(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(jsonAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(s.DocLines, jsonAnnotation.TypeStruct)
	return ok
}

func hasSlices(s intermediatemodel.Struct) bool {
	for _, f := range s.Fields {
		if f.IsSlice() {
			return true
		}
	}
	return false
}
