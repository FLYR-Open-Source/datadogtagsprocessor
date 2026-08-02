package traces

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
)

type Processor struct {
	mode     config.Mode
	contexts []config.ProcessorContext[ptrace.Traces]
	logger   *zap.Logger
}

func NewProcessor(mode config.Mode, contextStatements []config.ContextStatements, settings component.TelemetrySettings) (*Processor, error) {
	contexts := make([]config.ProcessorContext[ptrace.Traces], len(contextStatements))

	for i, cs := range contextStatements {
		contexts[i] = config.ProcessorContext[ptrace.Traces]{
			Consumer:          &traceStatements{},
			ContextStatements: cs,
		}
	}

	return &Processor{
		mode:     mode,
		contexts: contexts,
		logger:   settings.Logger,
	}, nil
}

func (p *Processor) ConsumeTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for _, c := range p.contexts {
		err := c.Consumer.Consume(ctx, td, p.mode, c.ContextStatements)
		if err != nil {
			p.logger.Error("failed processing traces", zap.Error(err))
			return td, err
		}
	}
	return td, nil
}

func (p *Processor) Shutdown(ctx context.Context) error {
	var errors error

	for _, c := range p.contexts {
		if shutdownable, ok := c.Consumer.(config.Shutdownable); ok {
			err := shutdownable.Shutdown(ctx)
			if err != nil {
				p.logger.Error("failed shutting down traces processor", zap.Error(err))
				errors = multierr.Append(errors, err)
			}
		}
	}
	return errors
}
