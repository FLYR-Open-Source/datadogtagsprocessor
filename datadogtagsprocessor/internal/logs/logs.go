package logs

import (
	"context"
	"slices"

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/validation"
)

type logStatements struct{}

func (*logStatements) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource, config.Log:
		return true
	default:
		return false
	}
}

func (*logStatements) Shutdown(ctx context.Context) error {
	return nil
}

func (*logStatements) Consume(ctx context.Context, plogs plog.Logs, cs config.ContextStatements) error {
	for i := 0; i < plogs.ResourceLogs().Len(); i++ {
		rlogs := plogs.ResourceLogs().At(i)
		resourceAttributes := rlogs.Resource().Attributes()

		var resourceAttributeKeys []string
		var resourceAttributeKeyValues []string
		if cs.Context == config.Resource {
			resourceAttributeKeys, resourceAttributeKeyValues = extraction.ExtractAttributeKeys(resourceAttributes, cs)
		}

		for j := 0; j < rlogs.ScopeLogs().Len(); j++ {
			slogs := rlogs.ScopeLogs().At(j)
			logs := slogs.LogRecords()

			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)
				logAttributes := log.Attributes()

				logAttributeKeys := []string{}
				logAttributeKeyValues := []string{}

				if cs.Context == config.Log {
					logAttributeKeys, logAttributeKeyValues = extraction.ExtractAttributeKeys(logAttributes, cs)
				}

				extraction.AddDDTags(logAttributes, slices.Concat(resourceAttributeKeyValues, logAttributeKeyValues))

				if cs.Mode == config.Move && cs.Context == config.Log {
					for _, key := range logAttributeKeys {
						logAttributes.Remove(key)
					}
				}
			}
		}

		if cs.Mode == config.Move && cs.Context == config.Resource && len(resourceAttributeKeys) > 0 {
			for _, key := range resourceAttributeKeys {
				resourceAttributes.Remove(key)
			}
		}
	}

	return nil
}

func NewLogParserCollection() *validation.ParserCollection[plog.Logs] {
	return validation.NewParserCollection(&logStatements{})
}
