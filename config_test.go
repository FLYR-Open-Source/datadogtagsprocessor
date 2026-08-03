package datadogtagsprocessor

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/metadata"
	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"
	"go.opentelemetry.io/collector/confmap/xconfmap"
)

func Test_LoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id       component.ID
		expected component.Config
		errors   []error
	}{
		{
			id: component.NewIDWithName(metadata.Type, ""),
			expected: &Config{
				TraceStatements: []config.ContextStatements{
					{
						Mode:       config.Merge,
						Context:    config.Resource,
						Attributes: []string{"k8s.*", "host.cpu.cache.l2.sizestring"},
					},
					{
						Mode:       config.Merge,
						Context:    config.Span,
						Attributes: []string{"team"},
					},
				},
				LogStatements: []config.ContextStatements{
					{
						Mode:       config.Merge,
						Context:    config.Resource,
						Attributes: []string{"k8s.*", "host.cpu.cache.l2.sizestring"},
					},
					{
						Mode:       config.Merge,
						Context:    config.Log,
						Attributes: []string{"team"},
					},
				},
			},
		},
		{
			id:       component.NewIDWithName(metadata.Type, "with_empty_context"),
			expected: nil,
			errors: []error{
				fmt.Errorf("context is empty for statements: %v", []string{"k8s.*", "host.cpu.cache.l2.sizestring"}),
				fmt.Errorf("context is empty for statements: %v", []string{"team"}),
			},
		},
		{
			id:       component.NewIDWithName(metadata.Type, "with_empty_attributes"),
			expected: nil,
			errors: []error{
				fmt.Errorf("empty list of attributes for context: \"resource\""),
				fmt.Errorf("empty list of attributes for context: \"log\""),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.id.Name(), func(t *testing.T) {
			cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
			require.NoError(t, err)

			factory := NewFactory()
			cfg := factory.CreateDefaultConfig()

			sub, err := cm.Sub(test.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			if test.expected == nil {
				err = xconfmap.Validate(cfg)
				assert.Error(t, err)

				if len(test.errors) > 0 {
					for _, expectedErr := range test.errors {
						assert.ErrorContains(t, err, expectedErr.Error())
					}
				}
			} else {
				require.NoError(t, xconfmap.Validate(cfg))
				assert.EqualExportedValues(t, test.expected, cfg)
			}
		})
	}
}
