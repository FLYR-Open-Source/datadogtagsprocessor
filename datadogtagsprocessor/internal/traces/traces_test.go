package traces

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

func TestTraceStatements_IsContextValid(t *testing.T) {
	tests := []struct {
		name     string
		context  config.ContextID
		expected bool
	}{
		{name: "resource is valid", context: config.Resource, expected: true},
		{name: "span is valid", context: config.Span, expected: true},
		{name: "log is invalid", context: config.Log, expected: false},
		{name: "empty is invalid", context: config.ContextID(""), expected: false},
	}

	stmts := &traceStatements{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, stmts.IsContextValid(test.context))
		})
	}
}

func TestTraceStatements_Shutdown(t *testing.T) {
	stmts := &traceStatements{}
	assert.NoError(t, stmts.Shutdown(t.Context()))
}

func buildTraces() ptrace.Traces {
	traces := ptrace.NewTraces()

	rspans := traces.ResourceSpans().AppendEmpty()
	rspans.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")

	sspans := rspans.ScopeSpans().AppendEmpty()

	span1 := sspans.Spans().AppendEmpty()
	span1.Attributes().PutStr("team", "payments")

	span2 := sspans.Spans().AppendEmpty()
	span2.Attributes().PutStr("team", "my-service")

	return traces
}

func TestTraceStatements_Consume_ResourceContext(t *testing.T) {
	traces := buildTraces()

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Merge, cs)
	assert.NoError(t, err)

	rspans := traces.ResourceSpans().At(0)

	// Resource-context statements tag the resource attributes themselves,
	// applied once per resource regardless of how many spans it has.
	ddtags, ok := rspans.Resource().Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())

	// Span attributes are untouched by a resource-context statement.
	for i := 0; i < rspans.ScopeSpans().At(0).Spans().Len(); i++ {
		_, ok := rspans.ScopeSpans().At(0).Spans().At(i).Attributes().Get("ddtags")
		assert.False(t, ok)
	}
}

func TestTraceStatements_Consume_SpanContext(t *testing.T) {
	traces := buildTraces()

	cs := config.ContextStatements{
		Context:    config.Span,
		Attributes: []string{"team"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Merge, cs)
	assert.NoError(t, err)

	sspans := traces.ResourceSpans().At(0).ScopeSpans().At(0)

	ddtags1, ok := sspans.Spans().At(0).Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:payments"}, ddtags1.Slice().AsRaw())

	ddtags2, ok := sspans.Spans().At(1).Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:my-service"}, ddtags2.Slice().AsRaw())
}

func TestTraceStatements_Consume_MultipleResources(t *testing.T) {
	traces := ptrace.NewTraces()

	rspansA := traces.ResourceSpans().AppendEmpty()
	rspansA.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")
	scopeSpansA := rspansA.ScopeSpans().AppendEmpty()
	spanA := scopeSpansA.Spans().AppendEmpty()
	spanA.Attributes().PutEmptySlice("noop")
	scopeSpansA.Spans().AppendEmpty().Attributes().PutEmptySlice("noop")

	rspansB := traces.ResourceSpans().AppendEmpty()
	rspansB.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-b")
	spanB := rspansB.ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	spanB.Attributes().PutEmptySlice("noop")

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Merge, cs)
	assert.NoError(t, err)

	ddtagsA, _ := rspansA.Resource().Attributes().Get("ddtags")
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtagsA.Slice().AsRaw())

	ddtagsB, _ := rspansB.Resource().Attributes().Get("ddtags")
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-b"}, ddtagsB.Slice().AsRaw())

	// Neither resource's tags leak onto the other's spans.
	_, ok := spanA.Attributes().Get("ddtags")
	assert.False(t, ok)
	_, ok = spanB.Attributes().Get("ddtags")
	assert.False(t, ok)
}

func TestTraceStatements_Consume_Move(t *testing.T) {
	traces := ptrace.NewTraces()
	span := traces.ResourceSpans().AppendEmpty().ScopeSpans().AppendEmpty().Spans().AppendEmpty()
	span.Attributes().PutStr("team", "payments")

	cs := config.ContextStatements{
		Context:    config.Span,
		Attributes: []string{"team"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Move, cs)
	assert.NoError(t, err)

	_, ok := span.Attributes().Get("team")
	assert.False(t, ok)

	ddtags, ok := span.Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:payments"}, ddtags.Slice().AsRaw())
}

func TestTraceStatements_Consume_ResourceContext_Move(t *testing.T) {
	traces := ptrace.NewTraces()

	rspans := traces.ResourceSpans().AppendEmpty()
	rspans.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")

	scopeSpans := rspans.ScopeSpans().AppendEmpty()
	scopeSpans.Spans().AppendEmpty().Attributes().PutEmptySlice("noop")
	scopeSpans.Spans().AppendEmpty().Attributes().PutEmptySlice("noop")
	scopeSpans.Spans().AppendEmpty().Attributes().PutEmptySlice("noop")

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Move, cs)
	assert.NoError(t, err)

	_, ok := rspans.Resource().Attributes().Get("k8s.cluster.name")
	assert.False(t, ok)

	ddtags, ok := rspans.Resource().Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())
}

func TestTraceStatements_Consume_NoResourceSpans(t *testing.T) {
	traces := ptrace.NewTraces()

	cs := config.ContextStatements{
		Context:    config.Span,
		Attributes: []string{"team"},
	}

	stmts := &traceStatements{}
	err := stmts.Consume(t.Context(), traces, config.Merge, cs)
	assert.NoError(t, err)
}

func TestNewTraceParserCollection(t *testing.T) {
	parser := NewTraceParserCollection()
	assert.NotNil(t, parser)

	err := parser.Parse(config.ContextStatements{
		Context:    config.Span,
		Attributes: []string{"team"},
	})
	assert.NoError(t, err)

	err = parser.Parse(config.ContextStatements{
		Context:    config.Log,
		Attributes: []string{"team"},
	})
	assert.Error(t, err)
}
