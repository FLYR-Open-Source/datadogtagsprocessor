package extraction

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestHandle(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("service.name", "api")

	cs := config.ContextStatements{
		Attributes: []string{"service.name"},
	}

	Handle(attributes, config.Merge, cs)

	ddtags, ok := attributes.Get(ddtagsKey)
	assert.True(t, ok)
	assert.ElementsMatch(t, []any{"service.name:api"}, ddtags.Slice().AsRaw())
}
