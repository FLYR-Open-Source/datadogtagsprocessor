package datadogtagsprocessor

import (
	"go.opentelemetry.io/collector/component"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/logs"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/traces"
)

// Config is the configuration of the datadog_tags processor.
type Config struct {
	// TraceStatements are the statements applied to traces.
	TraceStatements []config.ContextStatements `mapstructure:"trace_statements"`
	// LogStatements are the statements applied to logs.
	LogStatements []config.ContextStatements `mapstructure:"log_statements"`

	// logger is the processor logger.
	logger *zap.Logger
}

// Validate checks every configured statement at config load time.
//
// All statements are validated and the errors are collected, so a single
// run reports every invalid statement instead of just the first one.
func (c *Config) Validate() error {
	var errors error

	if len(c.TraceStatements) > 0 {
		parser := traces.NewTraceParserCollection()

		for _, cs := range c.TraceStatements {
			err := parser.Validate(cs)

			if err != nil {
				errors = multierr.Append(errors, err)
			}
		}
	}

	if len(c.LogStatements) > 0 {
		parser := logs.NewLogParserCollection()

		for _, cs := range c.LogStatements {
			err := parser.Validate(cs)

			if err != nil {
				errors = multierr.Append(errors, err)
			}
		}
	}

	return errors
}

var _ component.Config = (*Config)(nil)
