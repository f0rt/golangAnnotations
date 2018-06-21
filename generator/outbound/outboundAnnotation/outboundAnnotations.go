package outboundAnnotation

import "github.com/MarcGrol/golangAnnotations/generator/annotation"

const (
	TypeOutboundClient    = "OutboundClient"
	TypeOutboundOperation = "OutboundOperation"
	ParamName             = "name"
	ParamDescription      = "description"
	ParamExternal         = "external"
	ParamHttpBased        = "http"
	ParamInbound          = "inbound"
	ParamOutbound         = "outbound"
)

func Get() []annotation.AnnotationDescriptor {
	return []annotation.AnnotationDescriptor{
		{
			Name:       TypeOutboundClient,
			ParamNames: []string{ParamName, ParamDescription, ParamExternal, ParamHttpBased},
			Validator:  validateRestServiceAnnotation,
		},
		{
			Name:       TypeOutboundOperation,
			ParamNames: []string{ParamName, ParamDescription, ParamInbound, ParamOutbound},
			Validator:  validateRestOperationAnnotation,
		}}
}

func validateRestServiceAnnotation(annot annotation.Annotation) bool {
	if annot.Name == TypeOutboundClient {
		name, hasName := annot.Attributes[ParamDescription]
		return hasName && name != ""
	}
	return false
}

func validateRestOperationAnnotation(annot annotation.Annotation) bool {
	if annot.Name == TypeOutboundOperation {
		name, hasName := annot.Attributes[ParamDescription]
		return hasName && name != ""
	}
	return false
}
