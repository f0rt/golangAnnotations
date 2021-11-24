package event

import (
	"fmt"
	"log"
	"strings"
	"text/template"
	"unicode"

	generator "github.com/MarcGrol/golangAnnotations/codegeneration"
	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/event/eventAnnotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/generationUtil"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

type eventMap struct {
	Events          map[string]event
	IsAnyPersistent bool
	IsAnySensitive  bool
}

type event struct {
	Name         string
	IsPersistent bool
	IsSensitive  bool
}

type aggregateMap struct {
	PackageName  string
	AggregateMap map[string]eventMap
}

type structures struct {
	PackageName string
	Structs     []intermediatemodel.Struct
}

type Generator struct {
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func (eg *Generator) GetAnnotations() []annotation.AnnotationDescriptor {
	return eventAnnotation.Get()
}

func (eg *Generator) Generate(inputDir string, parsedSource intermediatemodel.ParsedSources) error {
	return generate(inputDir, parsedSource.Structs)
}

type generateContext struct {
	targetDir   string
	packageName string
	structs     []intermediatemodel.Struct
}

func generate(inputDir string, structs []intermediatemodel.Struct) error {
	packageName, err := generationUtil.GetPackageNameForStructs(structs)
	if packageName == "" || err != nil {
		return err
	}

	targetDir, err := generationUtil.DetermineTargetPath(inputDir, packageName)
	if err != nil {
		return err
	}

	ctx := generateContext{
		targetDir:   targetDir,
		packageName: packageName,
		structs:     structs,
	}

	err = generateAggregates(ctx)
	if err != nil {
		return err
	}

	err = generateWrappers(ctx)
	if err != nil {
		return err
	}

	err = generateAnonymized(ctx)
	if err != nil {
		return err
	}

	err = generateEventStore(ctx)
	if err != nil {
		return err
	}

	err = generateEventPublisher(ctx)
	if err != nil {
		return err
	}

	err = generateWrappersTest(ctx)
	if err != nil {
		return err
	}

	err = generateHandlerInterface(ctx)
	if err != nil {
		return err
	}

	return nil
}

func generateAggregates(ctx generateContext) error {

	aggregates := getAggregates(ctx.structs)
	if len(aggregates) == 0 {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/aggregates.go", ctx.targetDir)),
		TemplateName:   "aggregates",
		TemplateString: aggregateTemplate,
		FuncMap:        customTemplateFuncs,
		Data: aggregateMap{
			PackageName:  ctx.packageName,
			AggregateMap: aggregates,
		},
	})
	if err != nil {
		log.Fatalf("Error generating aggregates (%s)", err)
		return err
	}
	return nil
}

func getAggregates(structs []intermediatemodel.Struct) map[string]eventMap {
	aggregates := make(map[string]eventMap)
	eventCount := 0
	for _, s := range structs {
		if IsEvent(s) {
			events, ok := aggregates[GetAggregateName(s)]
			if !ok {
				events = eventMap{
					Events:          make(map[string]event),
					IsAnyPersistent: false,
					IsAnySensitive:  false,
				}
			}
			evt := event{
				Name:         s.Name,
				IsPersistent: IsPersistentEvent(s),
				IsSensitive:  IsSensitiveEvent(s),
			}
			if evt.IsPersistent {
				events.IsAnyPersistent = true
			}
			if evt.IsSensitive {
				events.IsAnySensitive = true
			}
			events.Events[s.Name] = evt
			aggregates[GetAggregateName(s)] = events
			eventCount++
		}
	}
	return aggregates
}

func generateWrappers(ctx generateContext) error {

	if !containsAny(ctx.structs, IsEvent) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/wrappers.go", ctx.targetDir)),
		TemplateName:   "wrappers",
		TemplateString: wrappersTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating wrappers for structures (%s)", err)
		return err
	}
	return nil
}

func generateAnonymized(ctx generateContext) error {

	if !containsAny(ctx.structs, IsSensitiveEventOrEventPart) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/anonymized.go", ctx.targetDir)),
		TemplateName:   "anonymized",
		TemplateString: anonymizedTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating anonymized for structures (%s)", err)
		return err
	}
	return nil
}

func containsAny(structs []intermediatemodel.Struct, predicate func(_ intermediatemodel.Struct) bool) bool {
	for _, s := range structs {
		if predicate(s) {
			return true
		}
	}
	return false
}

func generateEventStore(ctx generateContext) error {

	if !containsAny(ctx.structs, IsPersistentEvent) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/../%sStore/%sStore.go", ctx.targetDir, ctx.packageName, ctx.packageName)),
		TemplateName:   "event-store",
		TemplateString: eventStoreTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating event-store for structures (%s)", err)
		return err
	}
	return nil
}

func generateEventPublisher(ctx generateContext) error {

	if !containsAny(ctx.structs, isTransient) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/../%sPublisher/%sPublisher.go", ctx.targetDir, ctx.packageName, ctx.packageName)),
		TemplateName:   "event-publisher",
		TemplateString: eventPublisherTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating event-publisher for structures (%s)", err)
		return err
	}
	return nil
}

func generateWrappersTest(ctx generateContext) error {

	if !containsAny(ctx.structs, IsEvent) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/wrappers_test.go", ctx.targetDir)),
		TemplateName:   "wrappers-test",
		TemplateString: wrappersTestTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating wrappers-test for structures (%s)", err)
		return err
	}
	return nil
}

func generateHandlerInterface(ctx generateContext) error {

	if !containsAny(ctx.structs, IsEvent) {
		return nil
	}

	err := generationUtil.Generate(generationUtil.Info{
		Src:            ctx.packageName,
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/interface.go", ctx.targetDir)),
		TemplateName:   "interface",
		TemplateString: interfaceTemplate,
		FuncMap:        customTemplateFuncs,
		Data: structures{
			PackageName: ctx.packageName,
			Structs:     ctx.structs,
		},
	})
	if err != nil {
		log.Fatalf("Error generating interface for event-handlers (%s)", err)
		return err
	}
	return nil
}

var customTemplateFuncs = template.FuncMap{
	"GetEvents":                   GetEvents,
	"IsEvent":                     IsEvent,
	"IsRootEvent":                 IsRootEvent,
	"IsPersistentEvent":           IsPersistentEvent,
	"IsTransientEvent":            IsTransientEvent,
	"IsSensitiveEvent":            IsSensitiveEvent,
	"IsSensitiveEventOrEventPart": IsSensitiveEventOrEventPart,
	"IsSensitiveField":            IsSensitiveField,
	"IsDeepSensitiveField":        IsDeepSensitiveField,
	"IsCustomSensitiveField":      IsCustomSensitiveField,
	"GetAggregateName":            GetAggregateName,
	"GetAggregateNameLowerCase":   GetAggregateNameLowerCase,
	"EventIdentifier":             EventIdentifier,
	"FieldIdentifier":             FieldIdentifier,
	"SliceFieldIdentifier":        SliceFieldIdentifier,
	"HasValueForField":            hasValueForField,
	"ValueForField":               valueForField,
	"IsPointer":                   IsPointer,
	"IsSlice":                     IsSlice,
	"IsStringSlice":               IsStringSlice,
	"IsPrimitive":                 IsPrimitive,
	"IsBool":                      IsBool,
	"IsInt":                       IsInt,
	"IsString":                    IsString,
	"IsDate":                      IsDate,
	"IsCustom":                    IsCustom,
}

func GetEvents(thecontext structures) []intermediatemodel.Struct {
	eventsOnly := make([]intermediatemodel.Struct, 0)
	for _, s := range thecontext.Structs {
		if IsEvent(s) {
			eventsOnly = append(eventsOnly, s)
		}
	}
	return eventsOnly
}

func IsEvent(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(eventAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEvent)
	return ok
}

func IsEventPart(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(eventAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEventPart)
	return ok
}

func GetAggregateName(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(eventAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEvent); ok {
		return ann.Attributes[eventAnnotation.ParamAggregate]
	}
	return ""
}

func GetAggregateNameLowerCase(s intermediatemodel.Struct) string {
	return toFirstLower(GetAggregateName(s))
}

func IsRootEvent(s intermediatemodel.Struct) bool {
	if IsEvent(s) {
		annotations := annotation.NewRegistry(eventAnnotation.Get())
		if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEvent); ok {
			return ann.Attributes[eventAnnotation.ParamIsRootEvent] == "true"
		}
	}
	return false
}

func IsPersistentEvent(s intermediatemodel.Struct) bool {
	return IsEvent(s) && !isTransient(s)
}

func IsTransientEvent(s intermediatemodel.Struct) bool {
	return IsEvent(s) && isTransient(s)
}

func isTransient(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(eventAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEvent); ok {
		return ann.Attributes[eventAnnotation.ParamIsTransient] == "true"
	}
	return false
}

func IsSensitiveEventOrEventPart(s intermediatemodel.Struct) bool {
	return IsSensitiveEvent(s) || IsSensitiveEventPart(s)
}

func IsSensitiveEvent(s intermediatemodel.Struct) bool {
	if IsEvent(s) {
		annotations := annotation.NewRegistry(eventAnnotation.Get())
		if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEvent); ok {
			return ann.Attributes[eventAnnotation.ParamIsSensitive] == "true"
		}
	}
	return false
}

func IsSensitiveEventPart(s intermediatemodel.Struct) bool {
	if IsEventPart(s) {
		annotations := annotation.NewRegistry(eventAnnotation.Get())
		if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, eventAnnotation.TypeEventPart); ok {
			return ann.Attributes[eventAnnotation.ParamIsSensitive] == "true"
		}
	}
	return false
}

func IsSensitiveField(f intermediatemodel.Field) bool {
	return getSensitiveTag(f) == "true"
}

func IsDeepSensitiveField(f intermediatemodel.Field) bool {
	return getSensitiveTag(f) == "deep"
}

func IsCustomSensitiveField(f intermediatemodel.Field) bool {
	return getSensitiveTag(f) == "custom"
}

func getSensitiveTag(f intermediatemodel.Field) string {
	return f.GetTagMap()[eventAnnotation.FieldTagSensitive]
}

func hasValueForField(field intermediatemodel.Field) bool {
	if field.IsPrimitive() || field.IsPrimitiveSlice() {
		return true
	}
	return false
}

func valueForField(field intermediatemodel.Field) string {

	if field.IsInt() || field.IsIntSlice() {
		return valueForIntField(field)
	}

	if field.IsString() || field.IsStringSlice() {
		return valueForStringField(field)
	}

	if field.IsBool() {
		return valueForBoolField()
	}

	return ""
}

func valueForIntField(field intermediatemodel.Field) string {
	if field.IsSlice() {
		return "[]int{1,2}"
	}
	return "42"
}

func valueForStringField(field intermediatemodel.Field) string {
	if field.IsSlice() {
		return "[]string{" + fmt.Sprintf("\"Example1%s\"", field.Name) + "," +
			fmt.Sprintf("\"Example1%s\"", field.Name) + "}"
	}
	return fmt.Sprintf("\"Example3%s\"", field.Name)
}

func valueForBoolField() string {
	return "true"
}

func EventIdentifier(s intermediatemodel.Struct) string {
	if IsEvent(s) {
		return "e"
	}
	return "ep"
}

func FieldIdentifier(f intermediatemodel.Field) string {
	name := f.Name
	i := strings.LastIndex(name, ".")
	return toFirstLower(name[i+1:])
}

func SliceFieldIdentifier(f intermediatemodel.Field) string {
	name := f.SliceElementTypeName()
	i := strings.LastIndex(name, ".")
	return toFirstLower(name[i+1:])
}

func toFirstLower(in string) string {
	a := []rune(in)
	a[0] = unicode.ToLower(a[0])
	return string(a)
}

func IsPointer(f intermediatemodel.Field) bool {
	return f.IsPointer()
}

func IsSlice(f intermediatemodel.Field) bool {
	return f.IsSlice()
}

func IsStringSlice(f intermediatemodel.Field) bool {
	return f.IsStringSlice()
}

func IsPrimitive(f intermediatemodel.Field) bool {
	return f.IsPrimitive()
}

func IsBool(f intermediatemodel.Field) bool {
	return f.IsBool()
}

func IsInt(f intermediatemodel.Field) bool {
	return f.IsInt()
}

func IsString(f intermediatemodel.Field) bool {
	return f.IsString()
}

func IsDate(f intermediatemodel.Field) bool {
	return f.IsDate()
}

func IsCustom(f intermediatemodel.Field) bool {
	return f.IsCustom()
}
