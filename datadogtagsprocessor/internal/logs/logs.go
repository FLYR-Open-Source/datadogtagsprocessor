package logs

import (
	"context"

	"go.opentelemetry.io/collector/pdata/plog"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/extraction"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/validation"
)

type logStatements struct{}

func (*logStatements) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource, config.Log:
		return true
	default:
		return false
	}
}

func (*logStatements) Shutdown(ctx context.Context) error {
	return nil
}

func (*logStatements) Consume(ctx context.Context, plogs plog.Logs, mode config.Mode, cs config.ContextStatements) error {
	for i := 0; i < plogs.ResourceLogs().Len(); i++ {
		rlogs := plogs.ResourceLogs().At(i)

		if cs.Context == config.Resource {
			extraction.Handle(rlogs.Resource().Attributes(), mode, cs)
			continue
		}

		for j := 0; j < rlogs.ScopeLogs().Len(); j++ {
			slogs := rlogs.ScopeLogs().At(j)
			logs := slogs.LogRecords()

			for k := 0; k < logs.Len(); k++ {
				log := logs.At(k)
				extraction.Handle(log.Attributes(), mode, cs)
			}
		}
	}

	return nil
}

func NewLogParserCollection() *validation.ParserCollection[plog.Logs] {
	return validation.NewParserCollection(&logStatements{})
}
