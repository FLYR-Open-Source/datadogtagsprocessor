package traces

import (
	"context"

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

func (*traceStatements) Consume(ctx context.Context, ptraces ptrace.Traces, mode config.Mode, cs config.ContextStatements) error {
	for i := 0; i < ptraces.ResourceSpans().Len(); i++ {
		rspans := ptraces.ResourceSpans().At(i)

		if cs.Context == config.Resource {
			extraction.Handle(rspans.Resource().Attributes(), mode, cs)
			continue
		}

		for j := 0; j < rspans.ScopeSpans().Len(); j++ {
			sspans := rspans.ScopeSpans().At(j)
			spans := sspans.Spans()

			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				extraction.Handle(span.Attributes(), mode, cs)
			}
		}
	}
	return nil
}

func NewTraceParserCollection() *validation.ParserCollection[ptrace.Traces] {
	return validation.NewParserCollection(&traceStatements{})
}
