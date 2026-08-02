package logs

import (
	"context"
	"errors"
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/pdata/plog"
)

// mockConsumer implements config.Consumer[plog.Logs] without Shutdown,
// so it does not satisfy config.Shutdownable.
type mockConsumer struct {
	consumeErr error
}

func (*mockConsumer) IsContextValid(config.ContextID) bool { return true }

func (m *mockConsumer) Consume(context.Context, plog.Logs, config.Mode, config.ContextStatements) error {
	return m.consumeErr
}

// mockShutdownableConsumer additionally implements config.Shutdownable.
type mockShutdownableConsumer struct {
	mockConsumer
	shutdownErr error
}

func (m *mockShutdownableConsumer) Shutdown(context.Context) error {
	return m.shutdownErr
}

func newTelemetrySettings() component.TelemetrySettings {
	return componenttest.NewNopTelemetrySettings()
}

func TestNewProcessor(t *testing.T) {
	t.Run("with statements", func(t *testing.T) {
		statements := []config.ContextStatements{
			{Context: config.Resource, Attributes: []string{"k8s.cluster.name"}},
			{Context: config.Log, Attributes: []string{"team"}},
		}

		p, err := NewProcessor(config.Merge, statements, newTelemetrySettings())
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Len(t, p.contexts, 2)
	})

	t.Run("with no statements", func(t *testing.T) {
		p, err := NewProcessor(config.Merge, nil, newTelemetrySettings())
		assert.NoError(t, err)
		assert.NotNil(t, p)
		assert.Len(t, p.contexts, 0)
	})
}

func TestProcessor_ConsumeLogs(t *testing.T) {
	t.Run("applies all configured statements", func(t *testing.T) {
		statements := []config.ContextStatements{
			{Context: config.Resource, Attributes: []string{"k8s.cluster.name"}},
			{Context: config.Log, Attributes: []string{"team"}},
		}

		p, err := NewProcessor(config.Merge, statements, newTelemetrySettings())
		assert.NoError(t, err)

		logs := plog.NewLogs()
		rlogs := logs.ResourceLogs().AppendEmpty()
		rlogs.Resource().Attributes().PutStr("k8s.cluster.name", "cluster-a")
		log := rlogs.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
		log.Attributes().PutStr("team", "payments")

		result, err := p.ConsumeLogs(t.Context(), logs)
		assert.NoError(t, err)

		resourceDdtags, ok := result.ResourceLogs().At(0).Resource().Attributes().Get("ddtags")
		assert.True(t, ok)
		assert.ElementsMatch(t, []any{"k8s.cluster.name:cluster-a"}, resourceDdtags.Slice().AsRaw())

		logDdtags, ok := result.ResourceLogs().At(0).ScopeLogs().At(0).LogRecords().At(0).Attributes().Get("ddtags")
		assert.True(t, ok)
		assert.ElementsMatch(t, []any{"team:payments"}, logDdtags.Slice().AsRaw())
	})

	t.Run("stops and returns error on first failing consumer", func(t *testing.T) {
		p := &Processor{
			mode: config.Merge,
			contexts: []config.ProcessorContext[plog.Logs]{
				{Consumer: &mockConsumer{consumeErr: errors.New("boom")}, ContextStatements: config.ContextStatements{}},
			},
			logger: newTelemetrySettings().Logger,
		}

		logs := plog.NewLogs()
		result, err := p.ConsumeLogs(t.Context(), logs)
		assert.Error(t, err)
		assert.Equal(t, logs, result)
	})
}

func TestProcessor_Shutdown(t *testing.T) {
	t.Run("no contexts", func(t *testing.T) {
		p := &Processor{logger: newTelemetrySettings().Logger}
		assert.NoError(t, p.Shutdown(t.Context()))
	})

	t.Run("skips consumers that do not implement Shutdownable", func(t *testing.T) {
		p := &Processor{
			logger: newTelemetrySettings().Logger,
			contexts: []config.ProcessorContext[plog.Logs]{
				{Consumer: &mockConsumer{}},
			},
		}
		assert.NoError(t, p.Shutdown(t.Context()))
	})

	t.Run("propagates shutdown error", func(t *testing.T) {
		p := &Processor{
			logger: newTelemetrySettings().Logger,
			contexts: []config.ProcessorContext[plog.Logs]{
				{Consumer: &mockShutdownableConsumer{shutdownErr: errors.New("shutdown failed")}},
			},
		}

		err := p.Shutdown(t.Context())
		assert.Error(t, err)
		assert.ErrorContains(t, err, "shutdown failed")
	})

	t.Run("aggregates errors from multiple consumers", func(t *testing.T) {
		p := &Processor{
			logger: newTelemetrySettings().Logger,
			contexts: []config.ProcessorContext[plog.Logs]{
				{Consumer: &mockShutdownableConsumer{shutdownErr: errors.New("first")}},
				{Consumer: &mockShutdownableConsumer{}},
				{Consumer: &mockShutdownableConsumer{shutdownErr: errors.New("second")}},
			},
		}

		err := p.Shutdown(t.Context())
		assert.Error(t, err)
		assert.ErrorContains(t, err, "first")
		assert.ErrorContains(t, err, "second")
	})

	t.Run("real logStatements consumer has no-op shutdown", func(t *testing.T) {
		p, err := NewProcessor(config.Merge, []config.ContextStatements{
			{Context: config.Log, Attributes: []string{"team"}},
		}, newTelemetrySettings())
		assert.NoError(t, err)
		assert.NoError(t, p.Shutdown(t.Context()))
	})
}
