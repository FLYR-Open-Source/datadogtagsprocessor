package datadogtagsprocessor

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/metadata"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/processor/processortest"
)

func TestFactory_Type(t *testing.T) {
	factory := NewFactory()
	assert.Equal(t, factory.Type(), metadata.Type)
}

func TestFactoryCreateTraces(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	assert.Equal(t, component.StabilityLevelAlpha, factory.TracesStability())

	tp, err := factory.CreateTraces(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
	assert.NotNil(t, tp)
	assert.NoError(t, err)
}

func TestFactoryCreateLogs(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	assert.Equal(t, component.StabilityLevelAlpha, factory.LogsStability())

	lp, err := factory.CreateLogs(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
	assert.NotNil(t, lp)
	assert.NoError(t, err)
}

func TestFactoryCreateMetrics_ShouldFail(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()

	mp, err := factory.CreateMetrics(t.Context(), processortest.NewNopSettings(metadata.Type), cfg, consumertest.NewNop())
	assert.Nil(t, mp)
	assert.Error(t, err, "telemetry type is not supported")
}
