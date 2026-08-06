package config

import (
	"context"
	"fmt"
	"strings"
)

type Mode string

const (
	Move  Mode = "move"
	Merge Mode = "merge"
)

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

type ContextID string

const (
	Resource ContextID = "resource"
	Span     ContextID = "span"
	Log      ContextID = "log"
)

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

func (c *ContextID) String() string {
	return string(*c)
}

func (c *ContextID) IsEmpty() bool {
	return c.String() == ""
}

type Attribute string

type ContextStatements struct {
	Mode       Mode      `mapstructure:"mode"`
	Context    ContextID `mapstructure:"context"`
	Attributes []string  `mapstructure:"attributes"`
}

// CompiledAttribute is one entry of ContextStatements.Attributes with its
// wildcard syntax parsed. For a wildcard selection ("k8s.*"), Key holds the
// namespace ("k8s") and Prefix the ready-made match prefix ("k8s."); for an
// exact selection, Key holds the attribute key as written.
type CompiledAttribute struct {
	Key      string
	Prefix   string
	Wildcard bool
}

type CompiledStatement struct {
	Mode         Mode
	Context      ContextID
	Attributes   []CompiledAttribute
	HasWildcards bool
}

const wildcardSuffix = ".*"

// Compile parses the statement's attribute selections once so the per-record
// hot path never re-derives wildcard namespaces or prefixes. Attribute order
// is preserved, so tags are emitted in the order the config lists them.
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

type Consumer[T any] interface {
	IsContextValid(ContextID) bool
	Consume(context.Context, T, CompiledStatement) error
}

type Shutdownable interface {
	Shutdown(context.Context) error
}

type ProcessorContext[T any] struct {
	Consumer          Consumer[T]
	CompiledStatement CompiledStatement
}
