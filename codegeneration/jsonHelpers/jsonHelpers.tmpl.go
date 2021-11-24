package jsonHelpers

const jsonHelpersTemplate = `// Generated automatically by golangAnnotations: do not edit manually

package {{.PackageName}}

import (
	"encoding/json"
	"fmt"
)

{{range .Enums}}
{{$enum := .}}

// Helpers for json-enum {{.Name}}

var (
	_{{.Name}}NameToValue = map[string]{{.Name}}{
		{{range .EnumLiterals -}}
			"{{GetPreferredName $enum .}}": {{.Name}},
		{{end -}}
		{{if HasAlternativeName $enum -}}
			// alternative names for backward compatibility
			{{range .EnumLiterals -}}
				"{{GetAlternativeName $enum .}}": {{.Name}},
			{{end -}}
		{{end -}}
	}
	_{{.Name}}ValueToName = map[{{.Name}}]string{
		{{range .EnumLiterals -}}
			{{.Name}}: "{{GetPreferredName $enum .}}",
		{{end -}}
	}
)

{{if HasDefaultValue . -}}
	func {{.Name}}ByName(name string) {{.Name}} {
	t, ok := _{{.Name}}NameToValue[name]
	if !ok {
		return {{GetDefaultValue .}}
	}
	return t
}

{{end -}}

func (t {{.Name}}) String() string {
	v := _{{.Name}}ValueToName[t]
	return v
}

{{if not (HasAlternativeName $enum) -}}
func {{.Name}}EnumValues() []{{.Name}} {
	return []{{.Name}}{
	{{range .EnumLiterals -}}
		{{.Name}},
	{{end -}}
	}
}

func {{.Name}}EnumValuesAsString() []string {
	values := []string{}
	for _, e := range {{.Name}}EnumValues() {
		values = append(values, e.String())
	}
	return values
}

{{end -}}

// MarshalJSON caters for readable enums with a proper default value
func (t {{.Name}}) MarshalJSON() ([]byte, error) {
	s, ok := _{{.Name}}ValueToName[t]
	if !ok {
		{{if HasDefaultValue .}}s = _{{.Name}}ValueToName[{{GetDefaultValue .}}]{{else}}return nil, fmt.Errorf("invalid {{.Name}}: %d", t){{end}}
	}
	return json.Marshal(s)
}

// UnmarshalJSON caters for readable enums with a proper default value
func (t *{{.Name}}) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("{{.Name}} should be a string, got %s", data)
	}
	v, ok := _{{.Name}}NameToValue[s]
	if !ok {
		{{if HasDefaultValue .}}v = {{GetDefaultValue .}}{{else}}return fmt.Errorf("invalid {{.Name}} %q", s){{end}}
	}
	*t = v
	return nil
}

{{end -}}

{{range .Structs -}}

// Helpers for json-struct {{.Name}}
{{if HasSlices . -}}

// MarshalJSON prevents nil slices in json
func (data {{.Name}}) MarshalJSON() ([]byte, error) {
	type alias {{.Name}}
	var raw = alias(data)
	{{range .Fields -}}
		{{if .IsSlice -}}
			if raw.{{.Name}} == nil {
				raw.{{.Name}} = {{.TypeName}}{}
			}
		{{end -}}
	{{end -}}

	return json.Marshal(raw)
}

// UnmarshalJSON prevents nil slices from json
func (data *{{.Name}}) UnmarshalJSON(b []byte) error {
	type alias {{.Name}}
	var raw alias
	err := json.Unmarshal(b, &raw)

	{{range .Fields -}}
		{{if .IsSlice -}}
	if raw.{{.Name}} == nil {
		raw.{{.Name}} = {{.TypeName}}{}
	}
		{{end -}}
	{{end -}}

	*data = {{.Name}}(raw)

	return err
}

	{{end -}}
{{end -}}
`
