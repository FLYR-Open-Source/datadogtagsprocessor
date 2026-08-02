package datadogtagsprocessor

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/logs"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/metadata"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/traces"
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		metadata.Type,
		createDefaultConfig,
		processor.WithTraces(createTraces, metadata.TracesStability),
		processor.WithLogs(createLogs, metadata.LogsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createTraces(ctx context.Context, set processor.Settings, cfg component.Config, nextConsumer consumer.Traces) (processor.Traces, error) {
	oCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config for \"datadogtags\" processor.")
	}

	p, err := traces.NewProcessor(oCfg.TraceStatements, set.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("invalid config for \"transform\" processor %w", err)
	}

	return processorhelper.NewTraces(
		ctx,
		set,
		cfg,
		nextConsumer,
		p.ConsumeTraces,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}), // This is already the default, but we set it explicitly to make it clear that this processor mutates data.
		processorhelper.WithShutdown(p.Shutdown),
	)
}

func createLogs(ctx context.Context, set processor.Settings, cfg component.Config, nextConsumer consumer.Logs) (processor.Logs, error) {
	oCfg, ok := cfg.(*Config)
	if !ok {
		return nil, fmt.Errorf("invalid config for \"datadogtags\" processor.")
	}

	p, err := logs.NewProcessor(oCfg.LogStatements, set.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("invalid config for \"transform\" processor %w", err)
	}

	return processorhelper.NewLogs(
		ctx,
		set,
		cfg,
		nextConsumer,
		p.ConsumeLogs,
		processorhelper.WithCapabilities(consumer.Capabilities{MutatesData: true}), // This is already the default, but we set it explicitly to make it clear that this processor mutates data.
		processorhelper.WithShutdown(p.Shutdown),
	)
}
