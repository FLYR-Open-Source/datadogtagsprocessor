package validator

import (
	"fmt"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
)

// ParserCollection validates context statements for a signal type.
type ParserCollection[T any] struct {
	// consumer decides which contexts are valid for the signal.
	consumer config.Consumer[T]
}

// NewParserCollection builds a parser collection for the given consumer.
func NewParserCollection[T any](consumer config.Consumer[T]) *ParserCollection[T] {
	return &ParserCollection[T]{
		consumer: consumer,
	}
}

// Validate checks a statement at config load time.
//
// It rejects empty or unknown contexts, empty attribute lists and
// selections that can never match a real attribute.
func (p *ParserCollection[T]) Validate(cs config.ContextStatements) error {

	if cs.Context.IsEmpty() {
		return fmt.Errorf("context is empty for statements: %v", cs.Attributes)
	}

	if !p.consumer.IsContextValid(cs.Context) {
		return fmt.Errorf("unknown context %q for statements: %v", cs.Context.String(), cs.Attributes)
	}

	if len(cs.Attributes) == 0 {
		return fmt.Errorf("empty list of attributes for context: %q", cs.Context.String())
	}

	// A selection compiling to an empty key "" or a bare ".*" can never match a real attribute
	for i, attribute := range cs.Compile().Attributes {
		if attribute.Key == "" {
			return fmt.Errorf(
				"invalid attribute %q for context %q: missing attribute key",
				cs.Attributes[i], cs.Context.String(),
			)
		}
	}

	return nil
}
