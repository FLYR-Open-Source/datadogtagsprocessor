package logs

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
)

// Processor applies the configured log statements to incoming batches.
type Processor struct {
	// contexts holds one entry per configured statement, in config order.
	contexts []config.ProcessorContext[plog.Logs]
	// logger reports processing failures.
	logger *zap.Logger
}

// NewProcessor compiles the configured statements and builds a processor.
//
// Compilation happens once here so the hot path never parses the
// configuration again.
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

// ConsumeLogs runs every configured statement against the batch.
//
// Statements run in config order. The first failure stops processing and
// is returned.
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

// Shutdown shuts down every consumer that needs cleanup.
//
// All consumers are shut down and the errors are collected, so one failure
// does not stop the rest.
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
