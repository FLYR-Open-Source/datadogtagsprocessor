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

type Consumer[T any] interface {
	IsContextValid(ContextID) bool
	Consume(context.Context, T, ContextStatements) error
}

type Shutdownable interface {
	Shutdown(context.Context) error
}

type ProcessorContext[T any] struct {
	Consumer          Consumer[T]
	ContextStatements ContextStatements
}
