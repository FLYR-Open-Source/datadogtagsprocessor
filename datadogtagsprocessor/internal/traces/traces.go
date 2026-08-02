package traces

import (
	"context"
	"slices"

	"go.opentelemetry.io/collector/pdata/ptrace"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/validation"
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
			resourceAttributeKeys, resourceAttributeKeyValues = extraction.ExtractAttributeKeys(resourceAttributes, cs)
		}

		for j := 0; j < rspans.ScopeSpans().Len(); j++ {
			sspans := rspans.ScopeSpans().At(j)
			spans := sspans.Spans()

			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				spanAttributes := span.Attributes()

				spanAttributeKeys := []string{}
				spanAttributeKeyValues := []string{}

				if cs.Context == config.Span {
					spanAttributeKeys, spanAttributeKeyValues = extraction.ExtractAttributeKeys(spanAttributes, cs)
				}

				extraction.AddDDTags(spanAttributes, slices.Concat(resourceAttributeKeyValues, spanAttributeKeyValues))

				if cs.Mode == config.Move && cs.Context == config.Span {
					for _, key := range spanAttributeKeys {
						spanAttributes.Remove(key)
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

func NewTraceParserCollection() *validation.ParserCollection[ptrace.Traces] {
	return validation.NewParserCollection(&traceStatements{})
}
