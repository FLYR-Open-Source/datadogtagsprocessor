package logs

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
)

type Processor struct {
	contexts []config.ProcessorContext[plog.Logs]
	logger   *zap.Logger
}

func NewProcessor(contextStatements []config.ContextStatements, settings component.TelemetrySettings) (*Processor, error) {
	contexts := make([]config.ProcessorContext[plog.Logs], len(contextStatements))

	for i, cs := range contextStatements {
		contexts[i] = config.ProcessorContext[plog.Logs]{
			Consumer:          &logStatements{},
			CompiledStatement: cs.Compile(),
		}
	}

	return &Processor{
		contexts: contexts,
		logger:   settings.Logger,
	}, nil
}

func (p *Processor) ConsumeLogs(ctx context.Context, ld plog.Logs) (plog.Logs, error) {
	for _, c := range p.contexts {
		err := c.Consumer.Consume(ctx, ld, c.CompiledStatement)
		if err != nil {
			p.logger.Error("failed processing logs", zap.Error(err))
			return ld, err
		}
	}
	return ld, nil
}

func (p *Processor) Shutdown(ctx context.Context) error {
	var errors error

	for _, c := range p.contexts {
		if shutdownable, ok := c.Consumer.(config.Shutdownable); ok {
			err := shutdownable.Shutdown(ctx)
			if err != nil {
				p.logger.Error("failed shutting down log processor", zap.Error(err))
				errors = multierr.Append(errors, err)
			}
		}
	}
	return errors
}
