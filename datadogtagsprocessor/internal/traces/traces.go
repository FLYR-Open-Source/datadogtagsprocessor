package traces

import (
	"context"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/validator"
)

type traceStatements struct{}

func (*traceStatements) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource, config.Span:
		return true
	default:
		return false
	}
}

func (*traceStatements) Shutdown(ctx context.Context) error {
	return nil
}

func (*traceStatements) Consume(ctx context.Context, ptraces ptrace.Traces, cs config.ContextStatements) error {

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

		if cs.Mode == config.Move &&
			cs.Context == config.Resource {
			for _, key := range resourceAttributeKeys {
				resourceAttributes.Remove(key)
			}
		}
	}

	return nil
}

func NewTraceParserCollection() *validator.ParserCollection[ptrace.Traces] {
	return validator.NewParserCollection(&traceStatements{})
}
