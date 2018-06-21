package outbound

import (
	"log"

	"github.com/MarcGrol/golangAnnotations/generator"
	"github.com/MarcGrol/golangAnnotations/generator/annotation"
	"github.com/MarcGrol/golangAnnotations/generator/event/eventAnnotation"
	"github.com/MarcGrol/golangAnnotations/generator/generationUtil"
	"github.com/MarcGrol/golangAnnotations/generator/outbound/outboundAnnotation"
	"github.com/MarcGrol/golangAnnotations/model"
)

type Generator struct {
}

func NewGenerator() generator.Generator {
	return &Generator{}
}

func (eg *Generator) GetAnnotations() []annotation.AnnotationDescriptor {
	return eventAnnotation.Get()
}

func (eg *Generator) Generate(inputDir string, parsedSources model.ParsedSources) error {

	packageName, err := generationUtil.GetPackageNameForInterfaces(parsedSources.Interfaces)
	if err != nil {
		return err
	}
	targetDir, err := generationUtil.DetermineTargetPath(inputDir, packageName)
	if err != nil {
		return err
	}

	generateHttpClient(packageName, targetDir, parsedSources.Interfaces)

	return nil
}

func generateHttpClient(packageName, targetDir string, ifaces []model.Interface) {
	for _, iface := range ifaces {
		if isOutboundService(iface) {
			log.Printf("Outbound-client: %s.%s (%s)", packageName, iface.Name, getOutboundServiceDescription(iface))
			for _, o := range iface.Methods {
				if isOutboundOperation(o) {
					log.Printf("\tOutbound-operation: %s (%s)", o.Name, getOutboundOperationDescription(o))

				}
			}
		}
	}
}

func isOutboundService(iface model.Interface) bool {
	annotations := annotation.NewRegistry(outboundAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(iface.DocLines, outboundAnnotation.TypeOutboundClient)
	return ok
}

func isOutboundOperation(oper model.Operation) bool {
	annotations := annotation.NewRegistry(outboundAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(oper.DocLines, outboundAnnotation.TypeOutboundOperation)
	return ok
}

func getOutboundServiceDescription(iface model.Interface) string {
	annotations := annotation.NewRegistry(outboundAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(iface.DocLines, outboundAnnotation.TypeOutboundClient); ok {
		return ann.Attributes[outboundAnnotation.ParamDescription]
	}
	return ""
}

func getOutboundOperationDescription(oper model.Operation) string {
	annotations := annotation.NewRegistry(outboundAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(oper.DocLines, outboundAnnotation.TypeOutboundOperation); ok {
		return ann.Attributes[outboundAnnotation.ParamDescription]
	}
	return ""
}
