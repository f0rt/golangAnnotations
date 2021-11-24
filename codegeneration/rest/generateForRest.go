package rest

import (
	"fmt"
	"log"

	generator "github.com/MarcGrol/golangAnnotations/codegeneration"
	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/generationUtil"
	"github.com/MarcGrol/golangAnnotations/codegeneration/rest/restAnnotation"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

type Generator struct {
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func (eg *Generator) GetAnnotations() []annotation.AnnotationDescriptor {
	return restAnnotation.Get()
}

func (eg *Generator) Generate(inputDir string, parsedSource intermediatemodel.ParsedSources) error {
	return generate(inputDir, parsedSource.Structs)
}

type generateContext struct {
	targetDir   string
	packageName string
	service     intermediatemodel.Struct
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

	for _, service := range structs {
		if IsRestService(service) {
			ctx := generateContext{
				targetDir:   targetDir,
				packageName: packageName,
				service:     service,
			}
			err = generateHTTPService(ctx)
			if err != nil {
				return err
			}

			if !IsRestServiceNoTest(service) {
				err = generateHTTPTestHelpers(ctx)
				if err != nil {
					return err
				}
				err = generateHTTPTestService(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func generateHTTPService(ctx generateContext) error {
	err := generationUtil.Generate(generationUtil.Info{
		Src:            fmt.Sprintf("%s.%s", ctx.service.PackageName, ToFirstUpper(ctx.service.Name)),
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/http%s.go", ctx.targetDir, ToFirstUpper(ctx.service.Name))),
		TemplateName:   "http-handlers",
		TemplateString: httpHandlersTemplate,
		FuncMap:        customTemplateFuncs,
		Data:           ctx.service,
	})
	if err != nil {
		log.Fatalf("Error generating handlers for service %s: %s", ctx.service.Name, err)
		return err
	}
	return nil
}

func generateHTTPTestHelpers(ctx generateContext) error {
	err := generationUtil.Generate(generationUtil.Info{
		Src:            fmt.Sprintf("%s.%s", ctx.service.PackageName, ToFirstUpper(ctx.service.Name)),
		TargetFilename: generationUtil.Prefixed(fmt.Sprintf("%s/http%sHelpers_test.go", ctx.targetDir, ToFirstUpper(ctx.service.Name))),
		TemplateName:   "http-test-helpers",
		TemplateString: testHelpersTemplate,
		FuncMap:        customTemplateFuncs,
		Data:           ctx.service,
	})
	if err != nil {
		log.Fatalf("Error generating helpers for service %s: %s", ctx.service.Name, err)
		return err
	}
	return nil
}

func generateHTTPTestService(ctx generateContext) error {
	// create this file within a subdirectoty
	ctx.packageName = ctx.packageName + "TestLog"

	ctx.service.PackageName = ctx.packageName
	target := generationUtil.Prefixed(fmt.Sprintf("%s/%s/httpTest%s.go", ctx.targetDir, ctx.packageName, ToFirstUpper(ctx.service.Name)))
	err := generationUtil.Generate(generationUtil.Info{
		Src:            fmt.Sprintf("%s.%s", ctx.service.PackageName, ToFirstUpper(ctx.service.Name)),
		TargetFilename: target,
		TemplateName:   "testService",
		TemplateString: testServiceTemplate,
		FuncMap:        customTemplateFuncs,
		Data:           ctx.service,
	})
	if err != nil {
		log.Fatalf("Error generating testHandler for service %s: %s", ctx.service.Name, err)
		return err
	}
	return nil
}
