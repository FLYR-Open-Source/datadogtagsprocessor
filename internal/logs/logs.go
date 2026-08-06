package logs

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/validator"
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

func (*logStatements) Consume(ctx context.Context, plogs plog.Logs, cs config.CompiledStatement) error {
	for i := 0; i < plogs.ResourceLogs().Len(); i++ {
		rlogs := plogs.ResourceLogs().At(i)
		resourceAttributes := rlogs.Resource().Attributes()

		var resourceAttributeKeys []string
		var resourceAttributeKeyValues []string

		if cs.Context == config.Resource {
			resourceAttributeKeys, resourceAttributeKeyValues =
				extraction.ExtractAttributeKeys(resourceAttributes, cs)
		}

		for j := 0; j < rlogs.ScopeLogs().Len(); j++ {
			logs := rlogs.ScopeLogs().At(j).LogRecords()

			for k := 0; k < logs.Len(); k++ {
				extraction.ProcessRecordAttributes(
					logs.At(k).Attributes(),
					cs,
					resourceAttributeKeyValues,
				)
			}
		}

		if cs.Mode == config.Move && cs.Context == config.Resource {
			extraction.RemoveAttributes(resourceAttributes, resourceAttributeKeys)
		}
	}

	return nil
}

func NewLogParserCollection() *validator.ParserCollection[plog.Logs] {
	return validator.NewParserCollection(&logStatements{})
}
