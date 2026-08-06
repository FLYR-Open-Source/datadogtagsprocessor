package config

import (
	"context"
	"fmt"
	"strings"
)

// Mode selects what happens to the source attributes after they are copied
// into ddtags.
//
// Merge keeps the source attributes in place, while Move removes them once
// the tags are added.
type Mode string

const (
	// Move copies the selected attributes into ddtags and removes them
	// from the source.
	Move Mode = "move"
	// Merge copies the selected attributes into ddtags and keeps them
	// in the source.
	Merge Mode = "merge"
)

// UnmarshalText parses a mode from the collector configuration.
//
// The value is case insensitive. Anything other than "move" or "merge"
// fails the config load.
func (a *Mode) UnmarshalText(text []byte) error {
	str := Mode(strings.ToLower(string(text)))
	switch str {
	case Move, Merge:
		*a = str
		return nil
	default:
		return fmt.Errorf("unknown action %v", str)
	}
}

// ContextID is the level of the telemetry a statement runs against.
//
// Resource statements read the resource attributes, while span and log
// statements read the attributes of each individual record.
type ContextID string

const (
	// Resource applies a statement to the resource attributes.
	Resource ContextID = "resource"
	// Span applies a statement to the attributes of each span.
	Span ContextID = "span"
	// Log applies a statement to the attributes of each log record.
	Log ContextID = "log"
)

// UnmarshalText parses a context from the collector configuration.
//
// The value is case insensitive. Anything other than "resource", "span" or
// "log" fails the config load.
func (c *ContextID) UnmarshalText(text []byte) error {
	str := ContextID(strings.ToLower(string(text)))
	switch str {
	case Resource, Span, Log:
		*c = str
		return nil
	default:
		return fmt.Errorf("unknown context %v", str)
	}
}

// String returns the context as a plain string.
func (c *ContextID) String() string {
	return string(*c)
}

// IsEmpty reports whether the context was left unset in the configuration.
func (c *ContextID) IsEmpty() bool {
	return c.String() == ""
}

// Attribute is a single attribute selection as written in the configuration.
type Attribute string

// ContextStatements is a single statement block from the processor
// configuration.
//
// It describes which attributes to turn into ddtags, at which context and
// with which mode.
type ContextStatements struct {
	// Mode selects whether the source attributes are kept or removed.
	Mode Mode `mapstructure:"mode"`
	// Context selects the level the statement runs against.
	Context ContextID `mapstructure:"context"`
	// Attributes lists the attribute keys to select. A key ending in ".*"
	// selects every attribute under that namespace.
	Attributes []string `mapstructure:"attributes"`
}

// CompiledAttribute is one entry of ContextStatements.Attributes with its
// wildcard syntax parsed.
//
// For a wildcard selection like "k8s.*", Key holds the namespace ("k8s"),
// Prefix holds the ready-made match prefix ("k8s.") and Wildcard is true.
// For an exact selection, Key holds the attribute key as written.
type CompiledAttribute struct {
	// Key is the attribute key, or the namespace for a wildcard selection.
	Key string
	// Prefix is the namespace followed by a dot. It is only set for
	// wildcard selections.
	Prefix string
	// Wildcard reports whether the selection ends in ".*".
	Wildcard bool
}

// CompiledStatement is a ContextStatements with its attribute selections
// parsed once at processor start.
//
// The hot path works off this struct so it never has to re-parse the
// configured attributes per record.
type CompiledStatement struct {
	// Mode selects whether the source attributes are kept or removed.
	Mode Mode
	// Context selects the level the statement runs against.
	Context ContextID
	// Attributes holds the parsed selections in configuration order.
	Attributes []CompiledAttribute
	// HasWildcards reports whether at least one selection is a wildcard.
	HasWildcards bool
}

// wildcardSuffix marks a selection that matches a whole namespace.
const wildcardSuffix = ".*"

// Compile parses the statement's attribute selections.
//
// It runs once at processor start so the per-record hot path never
// re-derives wildcard namespaces or prefixes. Attribute order is preserved,
// so tags are emitted in the order the config lists them.
func (cs ContextStatements) Compile() CompiledStatement {
	compiled := CompiledStatement{
		Mode:       cs.Mode,
		Context:    cs.Context,
		Attributes: make([]CompiledAttribute, len(cs.Attributes)),
	}

	for i, attribute := range cs.Attributes {
		if namespace, ok := strings.CutSuffix(attribute, wildcardSuffix); ok {
			compiled.Attributes[i] = CompiledAttribute{
				Key:      namespace,
				Prefix:   namespace + ".",
				Wildcard: true,
			}
			compiled.HasWildcards = true
			continue
		}

		compiled.Attributes[i] = CompiledAttribute{Key: attribute}
	}

	return compiled
}

// Consumer applies a compiled statement to a batch of telemetry.
type Consumer[T any] interface {
	// IsContextValid reports whether the context is supported by this
	// signal type.
	IsContextValid(ContextID) bool
	// Consume applies the statement to every record in the batch.
	Consume(context.Context, T, CompiledStatement) error
}

// Shutdownable is implemented by consumers that need cleanup on shutdown.
type Shutdownable interface {
	// Shutdown releases any resources held by the consumer.
	Shutdown(context.Context) error
}

// ProcessorContext pairs a consumer with the compiled statement it runs.
type ProcessorContext[T any] struct {
	// Consumer applies the statement to the incoming telemetry.
	Consumer Consumer[T]
	// CompiledStatement is the statement given to the consumer.
	CompiledStatement CompiledStatement
}
