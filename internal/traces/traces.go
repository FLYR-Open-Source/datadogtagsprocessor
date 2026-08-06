package traces

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/validator"
)

// traceStatements applies compiled statements to trace batches.
type traceStatements struct{}

// IsContextValid reports whether the context is supported for traces.
//
// Only the resource and span contexts are valid.
func (*traceStatements) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource, config.Span:
		return true
	default:
		return false
	}
}

// Shutdown implements config.Shutdownable.
//
// There is nothing to clean up for traces.
func (*traceStatements) Shutdown(ctx context.Context) error {
	return nil
}

// Consume applies a statement to every span in the batch.
//
// For resource statements the tags are extracted once per resource and
// appended to every span. In move mode the matched resource attributes
// are removed after all spans are processed.
func (*traceStatements) Consume(ctx context.Context, ptraces ptrace.Traces, cs config.CompiledStatement) error {

	for i := 0; i < ptraces.ResourceSpans().Len(); i++ {
		rspans := ptraces.ResourceSpans().At(i)
		resourceAttributes := rspans.Resource().Attributes()

		var resourceAttributeKeys []string
		var resourceAttributeKeyValues []string

		if cs.Context == config.Resource {
			resourceAttributeKeys, resourceAttributeKeyValues =
				extraction.ExtractAttributeKeys(resourceAttributes, cs)
		}

		for j := 0; j < rspans.ScopeSpans().Len(); j++ {
			sspans := rspans.ScopeSpans().At(j).Spans()

			for k := 0; k < sspans.Len(); k++ {
				extraction.ProcessRecordAttributes(
					sspans.At(k).Attributes(),
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

// NewTraceParserCollection returns a parser collection that validates trace
// statements.
func NewTraceParserCollection() *validator.ParserCollection[ptrace.Traces] {
	return validator.NewParserCollection(&traceStatements{})
}
