package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModeUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected Mode
		wantErr  bool
	}{
		{name: "move", text: "move", expected: Move},
		{name: "merge", text: "merge", expected: Merge},
		{name: "uppercase is normalized", text: "MOVE", expected: Move},
		{name: "mixed case is normalized", text: "Merge", expected: Merge},
		{name: "unknown mode", text: "copy", wantErr: true},
		{name: "empty", text: "", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var mode Mode
			err := mode.UnmarshalText([]byte(test.text))

			if test.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, mode)
		})
	}
}

func TestContextIDUnmarshalText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected ContextID
		wantErr  bool
	}{
		{name: "resource", text: "resource", expected: Resource},
		{name: "span", text: "span", expected: Span},
		{name: "log", text: "log", expected: Log},
		{name: "uppercase is normalized", text: "RESOURCE", expected: Resource},
		{name: "mixed case is normalized", text: "Log", expected: Log},
		{name: "unknown context", text: "metric", wantErr: true},
		{name: "empty", text: "", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var context ContextID
			err := context.UnmarshalText([]byte(test.text))

			if test.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, test.expected, context)
		})
	}
}

func TestCompile(t *testing.T) {
	cs := ContextStatements{
		Mode:    Merge,
		Context: Resource,
		Attributes: []string{
			"k8s.*",
			"deployment.environment.name",
			"service.*",
			"host.name",
		},
	}

	compiled := cs.Compile()

	assert.Equal(t, CompiledStatement{
		Mode:    Merge,
		Context: Resource,
		Attributes: []CompiledAttribute{
			{Key: "k8s", Prefix: "k8s.", Wildcard: true},
			{Key: "deployment.environment.name"},
			{Key: "service", Prefix: "service.", Wildcard: true},
			{Key: "host.name"},
		},
		HasWildcards: true,
	}, compiled)
}

func TestCompile_WithoutWildcards(t *testing.T) {
	cs := ContextStatements{
		Mode:       Move,
		Context:    Log,
		Attributes: []string{"team", "host.name"},
	}

	compiled := cs.Compile()

	assert.Equal(t, CompiledStatement{
		Mode:    Move,
		Context: Log,
		Attributes: []CompiledAttribute{
			{Key: "team"},
			{Key: "host.name"},
		},
	}, compiled)
}

func TestCompile_WildcardParsing(t *testing.T) {
	tests := []struct {
		name      string
		attribute string
		expected  CompiledAttribute
	}{
		{
			name:      "wildcard suffix",
			attribute: "k8s.*",
			expected:  CompiledAttribute{Key: "k8s", Prefix: "k8s.", Wildcard: true},
		},
		{
			name:      "nested namespace wildcard",
			attribute: "k8s.cluster.*",
			expected:  CompiledAttribute{Key: "k8s.cluster", Prefix: "k8s.cluster.", Wildcard: true},
		},
		{
			name:      "no wildcard suffix",
			attribute: "k8s.cluster.name",
			expected:  CompiledAttribute{Key: "k8s.cluster.name"},
		},
		{
			name:      "wildcard not at the end is not stripped",
			attribute: "k8s.*.name",
			expected:  CompiledAttribute{Key: "k8s.*.name"},
		},
		{
			name:      "asterisk without dot prefix is not stripped",
			attribute: "k8s*",
			expected:  CompiledAttribute{Key: "k8s*"},
		},
		{
			name:      "just the wildcard marker",
			attribute: ".*",
			expected:  CompiledAttribute{Key: "", Prefix: ".", Wildcard: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compiled := ContextStatements{Attributes: []string{test.attribute}}.Compile()

			assert.Equal(t, []CompiledAttribute{test.expected}, compiled.Attributes)
			assert.Equal(t, test.expected.Wildcard, compiled.HasWildcards)
		})
	}
}
