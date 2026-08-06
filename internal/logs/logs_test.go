package logs

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/plog"
)

func buildLogs() plog.Logs {
	logs := plog.NewLogs()

	rlogs := logs.ResourceLogs().AppendEmpty()
	rlogs.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")

	slogs := rlogs.ScopeLogs().AppendEmpty()

	log1 := slogs.LogRecords().AppendEmpty()
	log1.Attributes().PutStr("team", "payments")

	log2 := slogs.LogRecords().AppendEmpty()
	log2.Attributes().PutStr("team", "my-service")

	return logs
}

func TestLogStatements_IsContextValid(t *testing.T) {
	tests := []struct {
		name     string
		context  config.ContextID
		expected bool
	}{
		{name: "resource is valid", context: config.Resource, expected: true},
		{name: "log is valid", context: config.Log, expected: true},
		{name: "span is invalid", context: config.Span, expected: false},
		{name: "empty is invalid", context: config.ContextID(""), expected: false},
	}

	stmts := &logStatements{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, stmts.IsContextValid(test.context))
		})
	}
}

func TestLogStatements_Shutdown(t *testing.T) {
	stmts := &logStatements{}
	assert.NoError(t, stmts.Shutdown(t.Context()))
}

func TestLogStatements_Consume_ResourceContext(t *testing.T) {
	logs := buildLogs()

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)

	rlogs := logs.ResourceLogs().At(0)

	_, ok := rlogs.Resource().Attributes().Get("ddtags")
	assert.False(t, ok)

	// Log record attributes are untouched by a resource-context statement.
	for i := 0; i < rlogs.ScopeLogs().At(0).LogRecords().Len(); i++ {
		ddtags, ok := rlogs.ScopeLogs().At(0).LogRecords().At(i).Attributes().Get("ddtags")
		assert.True(t, ok)
		assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())
	}
}

func TestLogStatements_Consume_LogContext(t *testing.T) {
	logs := buildLogs()

	cs := config.ContextStatements{
		Context:    config.Log,
		Attributes: []string{"team"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)

	slogs := logs.ResourceLogs().At(0).ScopeLogs().At(0)

	ddtags1, ok := slogs.LogRecords().At(0).Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:payments"}, ddtags1.Slice().AsRaw())

	ddtags2, ok := slogs.LogRecords().At(1).Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:my-service"}, ddtags2.Slice().AsRaw())
}

func TestLogStatements_Consume_MultipleResources(t *testing.T) {
	logs := plog.NewLogs()

	rlogsA := logs.ResourceLogs().AppendEmpty()
	rlogsA.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")
	scopeLogsA := rlogsA.ScopeLogs().AppendEmpty()
	logA := scopeLogsA.LogRecords().AppendEmpty()
	logA.Attributes().PutEmptySlice("noop")
	scopeLogsA.LogRecords().AppendEmpty().Attributes().PutEmptySlice("noop")

	rlogsB := logs.ResourceLogs().AppendEmpty()
	rlogsB.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-b")
	logB := rlogsB.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	logB.Attributes().PutEmptySlice("noop")

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)

	ddtagsA, _ := logA.Attributes().Get("ddtags")
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtagsA.Slice().AsRaw())

	ddtagsB, _ := logB.Attributes().Get("ddtags")
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-b"}, ddtagsB.Slice().AsRaw())

	// Neither resource's tags leak onto the other's log records.
	_, ok := rlogsA.Resource().Attributes().Get("ddtags")
	assert.False(t, ok)
	_, ok = rlogsB.Resource().Attributes().Get("ddtags")
	assert.False(t, ok)
}

func TestLogStatements_Consume_Move(t *testing.T) {
	logs := plog.NewLogs()
	log := logs.ResourceLogs().AppendEmpty().ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	log.Attributes().PutStr("team", "payments")

	cs := config.ContextStatements{
		Mode:       config.Move,
		Context:    config.Log,
		Attributes: []string{"team"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)

	_, ok := log.Attributes().Get("team")
	assert.False(t, ok)

	ddtags, ok := log.Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"team:payments"}, ddtags.Slice().AsRaw())
}

func TestLogStatements_Consume_ResourceContext_Move(t *testing.T) {
	logs := plog.NewLogs()

	rlogs := logs.ResourceLogs().AppendEmpty()
	rlogs.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")

	scopeLogs := rlogs.ScopeLogs().AppendEmpty()

	log1 := scopeLogs.LogRecords().AppendEmpty()
	log1.Attributes().PutEmptySlice("noop")

	log2 := scopeLogs.LogRecords().AppendEmpty()
	log2.Attributes().PutEmptySlice("noop")

	log3 := scopeLogs.LogRecords().AppendEmpty()
	log3.Attributes().PutEmptySlice("noop")

	cs := config.ContextStatements{
		Mode:       config.Move,
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.name"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)

	_, ok := rlogs.Resource().Attributes().Get("k8s.cluster.name")
	assert.False(t, ok)

	ddtags, ok := log1.Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())

	ddtags, ok = log2.Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())

	ddtags, ok = log3.Attributes().Get("ddtags")
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, ddtags.Slice().AsRaw())
}

func TestLogStatements_Consume_NoResourceLogs(t *testing.T) {
	logs := plog.NewLogs()

	cs := config.ContextStatements{
		Context:    config.Log,
		Attributes: []string{"team"},
	}

	stmts := &logStatements{}
	err := stmts.Consume(t.Context(), logs, cs.Compile())
	assert.NoError(t, err)
}

func TestNewLogParserCollection(t *testing.T) {
	parser := NewLogParserCollection()
	assert.NotNil(t, parser)

	err := parser.Validate(config.ContextStatements{
		Context:    config.Log,
		Attributes: []string{"team"},
	})
	assert.NoError(t, err)

	err = parser.Validate(config.ContextStatements{
		Context:    config.Span,
		Attributes: []string{"team"},
	})
	assert.Error(t, err)
}
