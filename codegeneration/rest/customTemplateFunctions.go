package rest

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/MarcGrol/golangAnnotations/codegeneration/annotation"
	"github.com/MarcGrol/golangAnnotations/codegeneration/rest/restAnnotation"
	"github.com/MarcGrol/golangAnnotations/intermediatemodel"
)

var customTemplateFuncs = template.FuncMap{
	"IsRestService":                         IsRestService,
	"ExtractImports":                        ExtractImports,
	"GetRestServicePath":                    GetRestServicePath,
	"GetExtractRequestContextMethod":        GetExtractRequestContextMethod,
	"DoesRestServiceRequireRoleValidation":  DoesRestServiceRequireRoleValidation,
	"IsRestOperation":                       IsRestOperation,
	"IsRestOperationNoWrap":                 IsRestOperationNoWrap,
	"IsRestOperationGenerated":              IsRestOperationGenerated,
	"HasRestOperationAfter":                 HasRestOperationAfter,
	"GetRestOperationPath":                  GetRestOperationPath,
	"GetRestOperationMethod":                GetRestOperationMethod,
	"IsRestOperationTransactional":          IsRestOperationTransactional,
	"IsRestOperationForm":                   IsRestOperationForm,
	"IsRestOperationJSON":                   IsRestOperationJSON,
	"IsRestOperationHTML":                   IsRestOperationHTML,
	"IsRestOperationCSV":                    IsRestOperationCSV,
	"IsRestOperationTXT":                    IsRestOperationTXT,
	"IsRestOperationMD":                     IsRestOperationMD,
	"IsRestOperationNoContent":              IsRestOperationNoContent,
	"IsRestOperationCustom":                 IsRestOperationCustom,
	"HasContentType":                        HasContentType,
	"GetContentType":                        GetContentType,
	"GetRestOperationFilename":              GetRestOperationFilename,
	"GetRestOperationRolesString":           GetRestOperationRolesString,
	"GetRestOperationProducesEvents":        GetRestOperationProducesEvents,
	"GetRestOperationProducesEventsAsSlice": GetRestOperationProducesEventsAsSlice,
	"HasOperationsWithInput":                HasOperationsWithInput,
	"HasInput":                              HasInput,
	"GetInputArgType":                       GetInputArgType,
	"GetOutputArgDeclaration":               GetOutputArgDeclaration,
	"GetOutputArgsDeclaration":              GetOutputArgsDeclaration,
	"GetOutputArgName":                      GetOutputArgName,
	"HasAnyPathParam":                       HasAnyPathParam,
	"IsSliceParam":                          IsSliceParam,
	"IsQueryParam":                          IsQueryParam,
	"GetInputArgName":                       GetInputArgName,
	"GetInputParamString":                   GetInputParamString,
	"GetOutputArgType":                      GetOutputArgType,
	"HasOutput":                             HasOutput,
	"HasMetaOutput":                         HasMetaOutput,
	"IsMetaCallback":                        IsMetaCallback,
	"IsIntArg":                              IsIntArg,
	"IsBoolArg":                             IsBoolArg,
	"IsStringArg":                           IsStringArg,
	"IsStringSliceArg":                      IsStringSliceArg,
	"IsDateArg":                             IsDateArg,
	"IsCustomArg":                           IsCustomArg,
	"RequiresParamValidation":               RequiresParamValidation,
	"IsInputArgMandatory":                   IsInputArgMandatory,
	"HasUpload":                             HasUpload,
	"IsUploadArg":                           IsUploadArg,
	"HasRequestContext":                     HasRequestContext,
	"HasContext":                            HasContext,
	"ReturnsError":                          ReturnsError,
	"NeedsContext":                          NeedsContext,
	"GetContextName":                        GetContextName,
	"WithBackTicks":                         SurroundWithBackTicks,
	"BackTick":                              BackTick,
	"ToFirstUpper":                          ToFirstUpper,
	"Uncapitalized":                         Uncapitalized,
}

func BackTick() string {
	return "`"
}

func SurroundWithBackTicks(body string) string {
	return fmt.Sprintf("`%s'", body)
}

func IsRestService(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService)
	return ok
}

func IsRestOperationTransactional(s intermediatemodel.Struct, o intermediatemodel.Operation) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamTransactional] == "true"
	}
	return false
}

func IsRestServiceUnprotected(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	ann, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService)
	return ok && ann.Attributes[restAnnotation.ParamProtected] != "true"
}

func GetRestServicePath(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService); ok {
		return ann.Attributes[restAnnotation.ParamPath]
	}
	return ""
}

func GetExtractRequestContextMethod(s intermediatemodel.Struct) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService); ok {
		switch ann.Attributes[restAnnotation.ParamCredentials] {
		case "all":
			return "request.NewContext"
		case "admin":
			return "request.NewAdminContext"
		case "none":
			return "request.NewMinimalContext"
		}
	}
	return "extractRequestContext"
}

func IsRestServiceNoValidation(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService); ok {
		return ann.Attributes[restAnnotation.ParamNoValidation] == "true"
	}
	return false
}

func DoesRestServiceRequireRoleValidation(s intermediatemodel.Struct) bool {
	return !IsRestServiceNoValidation(s)
}

func IsRestServiceNoTest(s intermediatemodel.Struct) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(s.DocLines, restAnnotation.TypeRestService); ok {
		return ann.Attributes[restAnnotation.ParamNoTest] == "true"
	}
	return false
}

func isImportToBeIgnored(imp string) bool {
	if imp == "" {
		return true
	}
	for _, i := range []string{
		"context",
		"github.com/gorilla/mux",
	} {
		if imp == i {
			return true
		}
	}
	return false
}

func ExtractImports(s intermediatemodel.Struct) []string {
	importsMap := map[string]bool{}
	for _, o := range s.Operations {
		for _, ia := range o.InputArgs {
			if !isImportToBeIgnored(ia.PackageName) {
				importsMap[ia.PackageName] = true
			}
		}
		for _, oa := range o.OutputArgs {
			if !isImportToBeIgnored(oa.PackageName) {
				importsMap[oa.PackageName] = true
			}
		}
	}
	return mapToSlice(importsMap)
}

func mapToSlice(importsMap map[string]bool) []string {
	importsList := make([]string, 0)
	for k := range importsMap {
		importsList = append(importsList, k)
	}
	return importsList
}

func HasOperationsWithInput(s intermediatemodel.Struct) bool {
	for _, o := range s.Operations {
		if HasInput(*o) == true {
			return true
		}
	}
	return false
}

func IsRestOperation(o intermediatemodel.Operation) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	_, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation)
	return ok
}

func IsRestOperationNoWrap(o intermediatemodel.Operation) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamNoWrap] == "true"
	}
	return false
}

func IsRestOperationGenerated(o intermediatemodel.Operation) bool {
	return !IsRestOperationNoWrap(o)
}

func HasRestOperationAfter(o intermediatemodel.Operation) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamAfter] == "true"
	}
	return false
}

func GetRestOperationPath(o intermediatemodel.Operation) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamPath]
	}
	return ""
}

func HasAnyPathParam(o intermediatemodel.Operation) bool {
	return len(getAllPathParams(o)) > 0
}

func getAllPathParams(o intermediatemodel.Operation) []string {
	re := regexp.MustCompile(`\{\w+\}`)
	path := GetRestOperationPath(o)
	params := re.FindAllString(path, -1)
	for idx, param := range params {
		params[idx] = param[1 : len(param)-1]
	}
	return params
}

func GetRestOperationMethod(o intermediatemodel.Operation) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamMethod]
	}
	return ""
}

func IsRestOperationForm(o intermediatemodel.Operation) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamForm] == "true"
	}
	return false
}

func GetRestOperationFormat(o intermediatemodel.Operation) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamFormat]
	}
	return ""
}

func IsRestOperationJSON(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "JSON"
}

func IsRestOperationHTML(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "HTML"
}

func IsRestOperationCSV(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "CSV"
}

func IsRestOperationTXT(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "TXT"
}

func IsRestOperationMD(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "MD"
}

func IsRestOperationNoContent(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "no_content"
}

func IsRestOperationCustom(o intermediatemodel.Operation) bool {
	return GetRestOperationFormat(o) == "custom"
}

func HasContentType(operation intermediatemodel.Operation) bool {
	return GetContentType(operation) != ""
}

func GetContentType(operation intermediatemodel.Operation) string {
	switch GetRestOperationFormat(operation) {
	case "JSON":
		return "application/json"
	case "HTML":
		return "text/html; charset=UTF-8"
	case "CSV":
		return "text/csv; charset=UTF-8"
	case "TXT":
		return "text/plain; charset=UTF-8"
	case "MD":
		return "text/markdown; charset=UTF-8"
	default:
		return ""
	}
}

func GetRestOperationFilename(o intermediatemodel.Operation) string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		return ann.Attributes[restAnnotation.ParamFilename]
	}
	return ""
}

func GetRestOperationRolesString(o intermediatemodel.Operation) string {
	roles := GetRestOperationRoles(o)
	for i, r := range roles {
		roles[i] = fmt.Sprintf("\"%s\"", r)
	}
	return fmt.Sprintf("[]string{%s}", strings.Join(roles, ","))
}

func GetRestOperationRoles(o intermediatemodel.Operation) []string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		if rolesAttr, ok := ann.Attributes[restAnnotation.ParamRoles]; ok {
			roles := strings.Split(rolesAttr, ",")
			for i, r := range roles {
				roles[i] = strings.Trim(r, " ")
			}
			return roles
		}
	}
	return []string{}
}

func GetRestOperationProducesEvents(o intermediatemodel.Operation) string {
	return asStringSlice(GetRestOperationProducesEventsAsSlice(o))
}

func GetRestOperationProducesEventsAsSlice(o intermediatemodel.Operation) []string {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	if ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation); ok {
		if attrs, ok := ann.Attributes[restAnnotation.ParamProducesEvents]; ok {
			eventsProduced := make([]string, 0)
			for _, e := range strings.Split(attrs, ",") {
				evt := strings.TrimSpace(e)
				if evt != "" {
					eventsProduced = append(eventsProduced, evt)
				}
			}
			return eventsProduced
		}
	}
	return []string{}
}

func asStringSlice(in []string) string {
	adjusted := make([]string, 0)
	for _, i := range in {
		adjusted = append(adjusted, fmt.Sprintf("\"%s\"", i))
	}
	return fmt.Sprintf("[]string{%s}", strings.Join(adjusted, ","))
}

func HasInput(o intermediatemodel.Operation) bool {
	if GetRestOperationMethod(o) == "POST" || GetRestOperationMethod(o) == "PUT" {
		for _, arg := range o.InputArgs {
			if IsInputArg(arg) {
				return true
			}
		}
	}
	return false
}

func HasRequestContext(o intermediatemodel.Operation) bool {
	for _, arg := range o.InputArgs {
		if IsRequestContextArg(arg) {
			return true
		}
	}
	return false
}

func HasContext(o intermediatemodel.Operation) bool {
	for _, arg := range o.InputArgs {
		if IsContextArg(arg) {
			return true
		}
	}
	return false
}

func ReturnsError(o intermediatemodel.Operation) bool {
	for _, arg := range o.OutputArgs {
		if IsErrorArg(arg) {
			return true
		}
	}
	return false
}

func NeedsContext(o intermediatemodel.Operation) bool {
	return HasContext(o) || ReturnsError(o)
}

func GetContextName(o intermediatemodel.Operation) string {
	for _, arg := range o.InputArgs {
		if IsContextArg(arg) {
			return arg.Name
		}
	}
	if ReturnsError(o) {
		return "c"
	}
	return ""
}

func GetInputArgType(o intermediatemodel.Operation) string {
	for _, arg := range o.InputArgs {
		if IsInputArg(arg) {
			return arg.DereferencedTypeName()
		}
	}
	return ""
}

func IsSliceParam(arg intermediatemodel.Field) bool {
	return arg.IsSlice()
}

func IsQueryParam(o intermediatemodel.Operation, arg intermediatemodel.Field) bool {
	if IsContextArg(arg) || IsRequestContextArg(arg) {
		return false
	}
	for _, pathParam := range getAllPathParams(o) {
		if pathParam == arg.Name {
			return false
		}
	}
	return true
}

func GetInputArgName(o intermediatemodel.Operation) string {
	for _, arg := range o.InputArgs {
		if IsInputArg(arg) {
			return arg.Name
		}
	}
	return ""
}

func GetInputParamString(o intermediatemodel.Operation) string {
	args := make([]string, 0)
	for _, arg := range o.InputArgs {
		args = append(args, arg.Name)
	}
	return strings.Join(args, ", ")
}

func HasOutput(o intermediatemodel.Operation) bool {
	for _, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			return true
		}
	}
	return false
}

func GetOutputArgType(o intermediatemodel.Operation) string {
	for _, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			return arg.TypeName
		}
	}
	return ""
}

func HasMetaOutput(o intermediatemodel.Operation) bool {
	return GetMetaArg(o) != nil
}

func IsMetaCallback(o intermediatemodel.Operation) bool {
	arg := GetMetaArg(o)
	return arg != nil && IsMetaCallbackArg(*arg)
}

func GetMetaArg(o intermediatemodel.Operation) *intermediatemodel.Field {
	var count = 0
	for _, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			count++
			if count == 2 {
				return &arg
			}
		}
	}
	return nil
}

func GetOutputArgDeclaration(o intermediatemodel.Operation) string {
	for _, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			return arg.EmptyInstance()
		}
	}
	return ""
}

func GetOutputArgsDeclaration(o intermediatemodel.Operation) []string {
	args := make([]string, 0)
	for idx, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			name := ""
			switch idx {
			case 0:
				name = "result"
			case 1:
				name = "meta"
			}
			args = append(args, fmt.Sprintf("var %s %s", name, arg.TypeName))
		}
	}
	return args
}

func GetOutputArgName(o intermediatemodel.Operation) string {
	for _, arg := range o.OutputArgs {
		if !IsErrorArg(arg) {
			if !arg.IsPointer() {
				return "&resp"
			}
			return "resp"
		}
	}
	return ""
}

func findArgInArray(array []string, toMatch string) bool {
	for _, p := range array {
		if strings.ToUpper(strings.TrimSpace(p)) == strings.ToUpper(toMatch) {
			return true
		}
	}
	return false
}

func RequiresParamValidation(o intermediatemodel.Operation) bool {
	for _, field := range o.InputArgs {
		if (IsIntArg(field) || IsBoolArg(field) || IsStringSliceArg(field) || IsStringArg(field)) && IsInputArgMandatory(o, field) {
			return true
		}
	}
	return false
}

func IsInputArgMandatory(o intermediatemodel.Operation, arg intermediatemodel.Field) bool {
	annotations := annotation.NewRegistry(restAnnotation.Get())
	ann, ok := annotations.ResolveAnnotationByName(o.DocLines, restAnnotation.TypeRestOperation)
	if !ok {
		return false
	}
	optionalArgsString, ok := ann.Attributes[restAnnotation.ParamOptional]
	if !ok {
		return true
	}

	return !findArgInArray(strings.Split(optionalArgsString, ","), arg.Name)
}

func HasUpload(o intermediatemodel.Operation) bool {
	for _, f := range o.InputArgs {
		if IsUploadArg(f) {
			return true
		}
	}
	return false
}

func IsInputArg(arg intermediatemodel.Field) bool {
	if IsCustomArg(arg) && !IsContextArg(arg) && !IsRequestContextArg(arg) {
		return true
	}
	return false
}

func IsErrorArg(f intermediatemodel.Field) bool {
	return f.TypeName == "error"
}

func IsUploadArg(f intermediatemodel.Field) bool {
	return f.Name == "upload"
}

func IsContextArg(f intermediatemodel.Field) bool {
	return f.TypeName == "context.Context"
}

func IsRequestContextArg(f intermediatemodel.Field) bool {
	return f.TypeName == "request.Context"
}

func IsMetaCallbackArg(f intermediatemodel.Field) bool {
	return f.TypeName == "errorh.MetaCallback"
}

func IsBoolArg(f intermediatemodel.Field) bool {
	return f.IsBool()
}

func IsIntArg(f intermediatemodel.Field) bool {
	return f.IsInt()
}

func IsStringArg(f intermediatemodel.Field) bool {
	return f.IsString()
}

func IsStringSliceArg(f intermediatemodel.Field) bool {
	return f.IsStringSlice()
}

func IsDateArg(f intermediatemodel.Field) bool {
	return f.IsDate()
}

func IsCustomArg(f intermediatemodel.Field) bool {
	return f.IsCustom()
}

func ToFirstUpper(in string) string {
	a := []rune(in)
	a[0] = unicode.ToUpper(a[0])
	return string(a)
}

func Uncapitalized(in string) string {
	out := make([]rune, 0)
	lastUpper := false
	for _, r := range []rune(in) {
		if unicode.IsUpper(r) {
			if lastUpper {
				r = unicode.ToLower(r)
			} else {
				lastUpper = true
			}
		} else {
			lastUpper = false
		}
		out = append(out, r)
	}
	return string(out)
}
