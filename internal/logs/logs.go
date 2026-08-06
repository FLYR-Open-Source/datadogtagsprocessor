package logs

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/validator"
)

// logStatements applies compiled statements to log batches.
type logStatements struct{}

// IsContextValid reports whether the context is supported for logs.
//
// Only the resource and log contexts are valid.
func (*logStatements) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource, config.Log:
		return true
	default:
		return false
	}
}

// Shutdown implements config.Shutdownable.
//
// There is nothing to clean up for logs.
func (*logStatements) Shutdown(ctx context.Context) error {
	return nil
}

// Consume applies a statement to every log record in the batch.
//
// For resource statements the tags are extracted once per resource and
// appended to every record. In move mode the matched resource attributes
// are removed after all records are processed.
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

// NewLogParserCollection returns a parser collection that validates log
// statements.
func NewLogParserCollection() *validator.ParserCollection[plog.Logs] {
	return validator.NewParserCollection(&logStatements{})
}
