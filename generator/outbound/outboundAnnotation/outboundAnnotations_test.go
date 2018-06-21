package outboundAnnotation

import (
	"testing"

	"github.com/MarcGrol/golangAnnotations/generator/annotation"
	"github.com/stretchr/testify/assert"
)

func TestCorrectOutboundtServiceAnnotation(t *testing.T) {
	registry := annotation.NewRegistry(Get())

	ann, ok := registry.ResolveAnnotationByName([]string{`// @OutboundClient( name = "myClient", description = "my descr", external = "true", http = "true")`}, "OutboundClient")
	assert.True(t, ok)
	assert.Equal(t, "myClient", ann.Attributes["name"])
	assert.Equal(t, "my descr", ann.Attributes["description"])
	assert.Equal(t, "true", ann.Attributes["http"])
	assert.Equal(t, "true", ann.Attributes["external"])
}

func TestIncompleteOutboundServiceAnnotation(t *testing.T) {
	registry := annotation.NewRegistry(Get())

	assert.Empty(t, registry.ResolveAnnotations([]string{`// @OutboundClient()`}))
}

func TestCorrectOutboundOperationAnnotation(t *testing.T) {
	registry := annotation.NewRegistry(Get())

	a, ok := registry.ResolveAnnotation(`// @OutboundOperation( name = "myOper", description = "my descr", inbound = "true", outbound = "false" )`)
	assert.True(t, ok)
	assert.Equal(t, "myOper", a.Attributes["name"])
	assert.Equal(t, "my descr", a.Attributes["description"])
	assert.Equal(t, "true", a.Attributes["inbound"])
	assert.Equal(t, "false", a.Attributes["outbound"])
}

func TestIncompleteRestOperationAnnotation(t *testing.T) {
	registry := annotation.NewRegistry(Get())

	assert.Empty(t, registry.ResolveAnnotations([]string{`// @OutboundClient()`}))
}
