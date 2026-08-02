package datadogtagsprocessor

import (
	"go.opentelemetry.io/collector/component"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/logs"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/traces"
)

type Config struct {
	TraceStatements []config.ContextStatements `mapstructure:"trace_statements"`
	LogStatements   []config.ContextStatements `mapstructure:"log_statements"`

	logger *zap.Logger
}

func (c *Config) Validate() error {
	var errors error

	if len(c.TraceStatements) > 0 {
		parser := traces.NewTraceParserCollection()

		for _, cs := range c.TraceStatements {
			err := parser.Parse(cs)

			if err != nil {
				errors = multierr.Append(errors, err)
			}
		}
	}

	if len(c.LogStatements) > 0 {
		parser := logs.NewLogParserCollection()

		for _, cs := range c.LogStatements {
			err := parser.Parse(cs)

			if err != nil {
				errors = multierr.Append(errors, err)
			}
		}
	}

	return errors
}

var _ component.Config = (*Config)(nil)
