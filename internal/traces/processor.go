package traces

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
)

// Processor applies the configured trace statements to incoming batches.
type Processor struct {
	// contexts holds one entry per configured statement, in config order.
	contexts []config.ProcessorContext[ptrace.Traces]
	// logger reports processing failures.
	logger *zap.Logger
}

// NewProcessor compiles the configured statements and builds a processor.
//
// Compilation happens once here so the hot path never parses the
// configuration again.
func NewProcessor(contextStatements []config.ContextStatements, settings component.TelemetrySettings) (*Processor, error) {
	contexts := make([]config.ProcessorContext[ptrace.Traces], len(contextStatements))

	for i, cs := range contextStatements {
		contexts[i] = config.ProcessorContext[ptrace.Traces]{
			Consumer:          &traceStatements{},
			CompiledStatement: cs.Compile(),
		}
	}

	return &Processor{
		contexts: contexts,
		logger:   settings.Logger,
	}, nil
}

// ConsumeTraces runs every configured statement against the batch.
//
// Statements run in config order. The first failure stops processing and
// is returned.
func (p *Processor) ConsumeTraces(ctx context.Context, td ptrace.Traces) (ptrace.Traces, error) {
	for _, c := range p.contexts {
		err := c.Consumer.Consume(ctx, td, c.CompiledStatement)
		if err != nil {
			p.logger.Error("failed processing traces", zap.Error(err))
			return td, err
		}
	}
	return td, nil
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
				p.logger.Error("failed shutting down traces processor", zap.Error(err))
				errors = multierr.Append(errors, err)
			}
		}
	}
	return errors
}
