package extraction

import (
	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func Handle(attributes pcommon.Map, mode config.Mode, cs config.ContextStatements) {
	extractAttributes(attributes, cs, mode)
}
