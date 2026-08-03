package validator

import (
	"context"
	"fmt"
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type mockConsumer struct{}

func (*mockConsumer) IsContextValid(context config.ContextID) bool {
	switch context {
	case config.Resource:
		return true
	default:
		return false
	}
}

func (l *mockConsumer) Consume(ctx context.Context, ptraces ptrace.Traces, cs config.ContextStatements) error {
	return nil
}

func newMockConsumer() *mockConsumer {
	return &mockConsumer{}
}

func newContextStatements(context string, attributes []string) config.ContextStatements {
	return config.ContextStatements{
		Context:    config.ContextID(context),
		Attributes: attributes,
	}
}

func TestNewParserCollection(t *testing.T) {
	tests := []struct {
		name       string
		context    string
		attributes []string
		error      error
	}{
		{
			name:       "Valid context and attributes",
			context:    "resource",
			attributes: []string{"attr1", "attr2"},
			error:      nil,
		},
		{
			name:       "Invalid context",
			context:    "span",
			attributes: []string{"attr1", "attr2"},
			error:      fmt.Errorf("unknown context %q for statements: %v", "span", []string{"attr1", "attr2"}),
		},
		{
			name:       "Empty context",
			context:    "",
			attributes: []string{"attr1", "attr2"},
			error:      fmt.Errorf("context is empty for statements: %v", []string{"attr1", "attr2"}),
		},
		{
			name:       "Empty attributes",
			context:    "resource",
			attributes: []string{},
			error:      fmt.Errorf("empty list of attributes for context: %q", "resource"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			consumer := newMockConsumer()
			parserCollection := NewParserCollection(consumer)

			err := parserCollection.Validate(newContextStatements(test.context, test.attributes))
			assert.Equal(t, test.error, err)
			assert.Equal(t, consumer, parserCollection.consumer)
		})
	}

}
