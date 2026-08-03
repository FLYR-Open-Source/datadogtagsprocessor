package validator

import (
	"fmt"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
)

type ParserCollection[T any] struct {
	consumer config.Consumer[T]
}

func NewParserCollection[T any](consumer config.Consumer[T]) *ParserCollection[T] {
	return &ParserCollection[T]{
		consumer: consumer,
	}
}

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

	return nil
}
