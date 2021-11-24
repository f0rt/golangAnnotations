package repository

const repositoryTemplate = `// Generated automatically by golangAnnotations: do not edit manually

package {{.PackageName}}

import (
	"context"

	"cloud.google.com/go/datastore"
)

{{if HasMethodFind . -}}
var Find{{UpperModelName .}}OnUID = DefaultFind{{UpperModelName .}}OnUID

func DefaultFind{{UpperModelName .}}OnUID(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string) (*{{ModelPackageName .}}.{{UpperModelName .}}, error) {
	{{LowerModelName .}}, _, err := DoFind{{UpperModelName .}}OnUID(c, rc, tx, {{LowerModelName .}}UID, envelope.AcceptAll)
	return {{LowerModelName .}}, err
}

{{if HasMethodFilterByEvent . -}}
func Find{{UpperModelName .}}OnUIDAndEvent(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string, metadata eventMetaData.Metadata) (*{{ModelPackageName .}}.{{UpperModelName .}}, error) {
	{{LowerModelName .}}, _, err := DoFind{{UpperModelName .}}OnUID(c, rc, tx, {{LowerModelName .}}UID, envelope.FilterByEventUID{EventUID: metadata.UUID})
	return {{LowerModelName .}}, err
}

{{end -}}

{{if HasMethodFilterByMoment . -}}
func Find{{UpperModelName .}}OnUIDAndMoment(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string, moment time.Time) (*{{ModelPackageName .}}.{{UpperModelName .}}, error) {
	{{LowerModelName .}}, _, err := DoFind{{UpperModelName .}}OnUID(c, rc, tx, {{LowerModelName .}}UID, envelope.FilterByMoment{Moment: moment})
	return {{LowerModelName .}}, err
}

{{end -}}

func DoFind{{UpperModelName .}}OnUID(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string, envelopeFilter envelope.EnvelopeFilter) (*{{ModelPackageName .}}.{{UpperModelName .}}, []envelope.Envelope, error) {
	envelopes, err := doFind{{UpperModelName .}}EnvelopesOnUID(c, rc, tx, {{LowerModelName .}}UID)
	if err != nil {
		return nil, nil, err
	}

	envelopes, err = envelopeFilter.FilteredEnvelopes(envelopes)
	if err != nil {
		return nil, nil, errorh.NewInternalErrorf(0, "Failed to filter events for {{LowerModelName .}} with uid %s: %s", {{LowerModelName .}}UID, err)
	}

	if len(envelopes) == 0 {
		return nil, nil, errorh.NewNotFoundErrorf(0, "{{UpperModelName .}} with uid %s not found", {{LowerModelName .}}UID)
	}

	{{LowerModelName .}} := {{ModelPackageName .}}.New{{UpperModelName .}}()
	err = {{GetPackageName .}}.Apply{{UpperAggregateName .}}Events(c, rc, envelopes, {{LowerModelName .}})
	if err != nil {
		return nil, nil, errorh.NewInternalErrorf(0, "Failed to apply %d events for {{LowerModelName .}} with uid %s: %s", len(envelopes), {{LowerModelName .}}UID, err)
	}
	return {{LowerModelName .}}, envelopes, nil
}

func doFind{{UpperModelName .}}EnvelopesOnUID(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string) ([]envelope.Envelope, error) {
	envelopes, err := eventStoreInstance.Search(c, rc, tx, {{GetPackageName .}}.{{AggregateNameConst .}}, {{LowerModelName .}}UID)
	if err != nil {
		return nil, errorh.NewInternalErrorf(0, "Failed to fetch events for {{LowerModelName .}} with uid %s: %s", {{LowerModelName .}}UID, err)
	}

	if len(envelopes) == 0 {
		return nil, errorh.NewNotFoundErrorf(0, "{{UpperModelName .}} with uid %s not found", {{LowerModelName .}}UID)
	}

	return envelopes, nil
}

{{end -}}

{{if HasMethodFindStates . -}}
	func Find{{UpperModelName .}}StatesOnUID(c context.Context, rc request.Context, tx *datastore.Transaction, {{LowerModelName .}}UID string) ([]{{ModelPackageName .}}.{{UpperModelName .}}, error) {
	envelopes, err := doFind{{UpperModelName .}}EnvelopesOnUID(c, rc, tx, {{LowerModelName .}}UID)
	if err != nil {
		return nil, err
	}

	states := make([]{{ModelPackageName .}}.{{UpperModelName .}}, 0, len(envelopes))
	{{LowerModelName .}} := {{ModelPackageName .}}.New{{UpperModelName .}}()
	for _, envlp := range envelopes {
		err = {{GetPackageName .}}.Apply{{UpperAggregateName .}}Event(c, rc, envlp, {{LowerModelName .}})
		if err != nil {
			return nil, errorh.NewInternalErrorf(0, "Failed to apply '%s' for {{LowerModelName .}} with uid %s: %s", envlp.EventTypeName, {{LowerModelName .}}UID, err)
		}
		states = append(states, *{{LowerModelName .}})
	}
	return states, nil
	}

{{end -}}

{{if HasMethodExists . -}}
func Exists{{UpperModelName .}}OnUID(c context.Context, rc request.Context, {{LowerModelName .}}UID string) (bool, error) {
	exists, err := eventStoreInstance.Exists(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}}, {{LowerModelName .}}UID)
	if err != nil {
		return false, errorh.NewInternalErrorf(0, "Failed to fetch events for {{LowerModelName .}} with uid %s: %s", {{LowerModelName .}}UID, err)
	}
	return exists, nil
}

{{end -}}

{{if HasMethodAllAggregateUIDs . -}}
func GetAll{{UpperModelName .}}UIDs(c context.Context, rc request.Context) ([]string, error) {
	{{LowerModelName .}}UIDs, err := eventStoreInstance.GetAllAggregateUIDs(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}})
	if err != nil {
		return nil, errorh.NewInternalErrorf(0, "Failed to fetch all {{LowerModelName .}} uids: %s", err)
	}
		return {{LowerModelName .}}UIDs, nil
	}

{{end -}}

{{if HasMethodGetAllAggregates . -}}
func GetAllRecent{{UpperModelName .}}s(c context.Context, rc request.Context, optOffset time.Time) ([]{{ModelPackageName .}}.{{UpperModelName .}}, error) {
			{{LowerModelName .}}, _, err := DoGetAllRecent{{UpperModelName .}}s(c, rc, optOffset)
	return {{LowerModelName .}}, err
}

func DoGetAllRecent{{UpperModelName .}}s(c context.Context, rc request.Context, optOffset time.Time) ([]{{ModelPackageName .}}.{{UpperModelName .}}, map[string][]envelope.Envelope, error) {
			{{LowerModelName .}}Map := map[string][]envelope.Envelope{}
	err := eventStoreInstance.IterateWithOffset(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}}, optOffset, func(envlp envelope.Envelope) error {
		envelopes, exists := {{LowerModelName .}}Map[envlp.AggregateUID]
		if !exists {
			envelopes = []envelope.Envelope{}
		}
		{{LowerModelName .}}Map[envlp.AggregateUID] = append(envelopes, envlp)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	{{LowerModelName .}}s := make([]{{ModelPackageName .}}.{{UpperModelName .}}, 0, len({{LowerModelName .}}Map))
	for _, {{LowerAggregateName .}}Envelopes := range {{LowerModelName .}}Map {
		// Sort events of aggregate on order of arrival (because appengine returns undeterministic order)
		sort.Slice({{LowerAggregateName .}}Envelopes, func(i, j int) bool {
			return {{LowerAggregateName .}}Envelopes[i].Timestamp.Before({{LowerAggregateName .}}Envelopes[j].Timestamp)
		})

		{{LowerModelName .}} := {{ModelPackageName .}}.New{{UpperModelName .}}()
		{{GetPackageName .}}.Apply{{UpperAggregateName .}}Events(c, rc, {{LowerAggregateName .}}Envelopes, {{LowerModelName .}})
		{{LowerModelName .}}s = append({{LowerModelName .}}s, *{{LowerModelName .}})
	}
	return {{LowerModelName .}}s, {{LowerModelName .}}Map, nil
}

{{end -}}

{{if HasMethodPurgeOnEventUIDs . -}}
	func Purge{{UpperAggregateName .}}EnvelopesOnUID(c context.Context, rc request.Context, {{LowerModelName .}}UID string, eventUUIDs []string) error {
	return eventStoreInstance.Purge(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}}, {{LowerModelName .}}UID, eventUUIDs)
}

{{end -}}

{{if HasMethodPurgeOnEventType . -}}
func PurgeAll{{UpperAggregateName .}}EnvelopesOnEventType(c context.Context, rc request.Context, eventTypeName string) (bool, error) {
	if eventTypeName == "" {
	return false, errorh.NewInvalidInputErrorf(0, "Missing eventTypeName")
	}
	done, err := eventStoreInstance.PurgeAll(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}}, eventTypeName)
	if err != nil {
	return false, errorh.NewInternalErrorf(0, "Failed to purge all '%s/%s' events: %s", {{GetPackageName .}}.{{AggregateNameConst .}}, eventTypeName, err)
	}
	return done, nil
}

{{end -}}

{{if HasMethodPurgeAll . -}}
func PurgeAll{{UpperAggregateName .}}Envelopes(c context.Context, rc request.Context) (bool, error) {
	done, err := eventStoreInstance.PurgeAll(c, rc, {{GetPackageName .}}.{{AggregateNameConst .}}, "")
	if err != nil {
	return false, errorh.NewInternalErrorf(0, "Failed to purge all '%s' events: %s", {{GetPackageName .}}.{{AggregateNameConst .}}, err)
	}
	return done, nil
}

{{end -}}
`
