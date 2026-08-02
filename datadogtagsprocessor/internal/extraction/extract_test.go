package extraction

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestHasWildcardSuffix(t *testing.T) {
	tests := []struct {
		name              string
		attribute         string
		expectedNamespace string
		expectedOK        bool
	}{
		{name: "wildcard suffix", attribute: "k8s.*", expectedNamespace: "k8s", expectedOK: true},
		{
			name:              "nested namespace wildcard",
			attribute:         "k8s.cluster.*",
			expectedNamespace: "k8s.cluster",
			expectedOK:        true,
		},
		{name: "no wildcard suffix", attribute: "k8s.cluster.name", expectedNamespace: "k8s.cluster.name", expectedOK: false},
		{name: "empty string", attribute: "", expectedNamespace: "", expectedOK: false},
		{name: "just the wildcard marker", attribute: ".*", expectedNamespace: "", expectedOK: true},
		{
			name:              "wildcard not at the end is not stripped",
			attribute:         "k8s.*.name",
			expectedNamespace: "k8s.*.name",
			expectedOK:        false,
		},
		{
			name:              "asterisk without dot prefix is not stripped",
			attribute:         "k8s*",
			expectedNamespace: "k8s*",
			expectedOK:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			namespace, ok := hasWildcardSuffix(test.attribute)

			assert.Equal(t, test.expectedOK, ok)
			assert.Equal(t, test.expectedNamespace, namespace)
		})
	}
}

func TestWildcardNamespaces(t *testing.T) {
	tests := []struct {
		name       string
		attributes []string
		expected   []string
	}{
		{name: "no attributes", attributes: []string{}, expected: nil},
		{name: "no wildcard", attributes: []string{"service.name", "team"}, expected: nil},
		{name: "wildcard present", attributes: []string{"service.name", "k8s.*"}, expected: []string{"k8s"}},
		{name: "only wildcard", attributes: []string{"k8s.*"}, expected: []string{"k8s"}},
		{
			name:       "multiple wildcards",
			attributes: []string{"k8s.*", "team", "http.*"},
			expected:   []string{"k8s", "http"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, wildcardNamespaces(test.attributes))
		})
	}
}

func TestBuildAttributeLookup(t *testing.T) {
	attributes := pcommon.NewMap()

	attributes.PutStr("k8s.cluster.name", "cluster")
	attributes.PutStr("k8s.cluster.uid", "123")
	attributes.PutStr("k8s.pod.name", "pod")

	root := buildAttributeLookup(attributes, []string{"k8s"})

	k8s := root.children["k8s"]
	cluster := k8s.children["cluster"]
	pod := k8s.children["pod"]

	assert.Contains(t, root.children, "k8s")
	assert.Contains(t, k8s.children, "cluster")
	assert.Contains(t, k8s.children, "pod")

	assert.Contains(t, cluster.children, "name")
	assert.Contains(t, cluster.children, "uid")
	assert.Contains(t, pod.children, "name")

	assert.True(t, cluster.children["name"].isLeaf)
	assert.True(t, cluster.children["uid"].isLeaf)
	assert.True(t, pod.children["name"].isLeaf)
}

func TestBuildAttributeLookup_OnlyRequestedNamespaces(t *testing.T) {
	attributes := pcommon.NewMap()

	attributes.PutStr("k8s.cluster.name", "cluster")
	attributes.PutStr("http.method", "GET")
	attributes.PutStr("team", "payments")

	root := buildAttributeLookup(attributes, []string{"k8s"})

	assert.Contains(t, root.children, "k8s")
	assert.NotContains(t, root.children, "http")
	assert.NotContains(t, root.children, "team")
}

func TestBuildAttributeLookup_DoesNotMatchNamespacePrefixOfLongerKey(t *testing.T) {
	attributes := pcommon.NewMap()

	attributes.PutStr("k8s.pod.name", "pod")
	attributes.PutStr("k8s.podname", "other")

	root := buildAttributeLookup(attributes, []string{"k8s.pod"})

	pod := root.children["k8s"].children["pod"]

	assert.ElementsMatch(t, []string{"k8s.pod.name"}, pod.keys)
}

func TestLookupSelectedAttributes(t *testing.T) {
	attributes := pcommon.NewMap()

	attributes.PutStr("k8s.cluster.name", "cluster")
	attributes.PutStr("k8s.cluster.uid", "123")
	attributes.PutStr("k8s.pod.name", "pod")

	root := buildAttributeLookup(attributes, []string{"k8s"})

	tests := []struct {
		name     string
		selected string
		expected []string
	}{
		{
			name:     "wildcard namespace",
			selected: "k8s.cluster.*",
			expected: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
			},
		},
		{
			name:     "exact attribute",
			selected: "k8s.pod.name",
			expected: []string{
				"k8s.pod.name",
			},
		},
		{
			name:     "non-existing attribute",
			selected: "k8s.service.name",
			expected: nil,
		},
		{
			name:     "non-existing namespace",
			selected: "k8s.service.*",
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := lookupSelectedAttributes(root, attributes, test.selected)

			assert.ElementsMatch(t, test.expected, result)
		})
	}
}

func TestGetDDTags(t *testing.T) {
	t.Run("should create ddtags when it does not exist", func(t *testing.T) {
		attributes := pcommon.NewMap()

		getDDTags(attributes)

		ddtags, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Empty(t, ddtags.Slice().AsRaw())
	})

	t.Run("should not overwrite existing ddtags", func(t *testing.T) {
		attributes := pcommon.NewMap()

		ddtags := attributes.PutEmptySlice(ddtagsKey)
		ddtags.AppendEmpty().SetStr("existing:value")

		getDDTags(attributes)

		value, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Equal(t, []any{"existing:value"}, value.Slice().AsRaw())
	})
}

func TestAddDDTag(t *testing.T) {
	attributes := pcommon.NewMap()

	addDDTags(attributes, []string{"service.name:my-service"})

	ddtags, ok := attributes.Get(ddtagsKey)

	assert.True(t, ok)
	assert.Equal(
		t,
		[]any{"service.name:my-service"},
		ddtags.Slice().AsRaw(),
	)
}

func TestExtractAttributeKeys(t *testing.T) {
	t.Run("exact attribute", func(t *testing.T) {
		attributes := pcommon.NewMap()
		attributes.PutStr("service.name", "my-service")
		attributes.PutStr("service.version", "1.2.3")

		cs := config.ContextStatements{
			Attributes: []string{"service.name"},
		}

		keys, values := ExtractAttributeKeys(attributes, cs)

		assert.ElementsMatch(t, []string{"service.name"}, keys)
		assert.ElementsMatch(t, []string{"service.name:my-service"}, values)

		// Resolving keys must not mutate the source attributes.
		_, ok := attributes.Get("service.name")
		assert.True(t, ok)
	})

	t.Run("wildcard namespace", func(t *testing.T) {
		attributes := pcommon.NewMap()
		attributes.PutStr("k8s.cluster.name", "cluster")
		attributes.PutStr("k8s.cluster.uid", "123")
		attributes.PutStr("k8s.pod.name", "pod")

		cs := config.ContextStatements{
			Attributes: []string{"k8s.cluster.*"},
		}

		keys, values := ExtractAttributeKeys(attributes, cs)

		assert.ElementsMatch(t, []string{"k8s.cluster.name", "k8s.cluster.uid"}, keys)
		assert.ElementsMatch(
			t,
			[]string{"k8s.cluster.name:cluster", "k8s.cluster.uid:123"},
			values,
		)
	})

	t.Run("non-existing attribute yields no keys or values", func(t *testing.T) {
		attributes := pcommon.NewMap()
		attributes.PutStr("service.name", "my-service")

		cs := config.ContextStatements{
			Attributes: []string{"service.missing"},
		}

		keys, values := ExtractAttributeKeys(attributes, cs)

		assert.Empty(t, keys)
		assert.Empty(t, values)
	})
}

func TestExtractAttributes_Merge(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("service.name", "my-service")
	attributes.PutStr("service.version", "1.2.3")

	cs := config.ContextStatements{
		Attributes: []string{"service.name"},
	}

	_, values := ExtractAttributeKeys(attributes, cs)
	addDDTags(attributes, values)

	value, ok := attributes.Get("service.name")
	assert.True(t, ok)

	assert.Equal(t, "my-service", value.AsString())

	ddtags, ok := attributes.Get(ddtagsKey)
	assert.True(t, ok)

	assert.ElementsMatch(
		t,
		[]any{"service.name:my-service"},
		ddtags.Slice().AsRaw(),
	)
}

func TestExtractAttributes_Move(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("service.name", "my-service")
	attributes.PutStr("service.version", "1.2.3")

	cs := config.ContextStatements{
		Attributes: []string{"service.name"},
	}

	keys, values := ExtractAttributeKeys(attributes, cs)
	addDDTags(attributes, values)

	for _, key := range keys {
		attributes.Remove(key)
	}

	_, exists := attributes.Get("service.name")
	assert.False(t, exists)

	value, ok := attributes.Get("service.version")
	assert.True(t, ok)
	assert.Equal(t, "1.2.3", value.AsString())

	ddtags, ok := attributes.Get(ddtagsKey)
	assert.True(t, ok)

	assert.ElementsMatch(
		t,
		[]any{"service.name:my-service"},
		ddtags.Slice().AsRaw(),
	)
}

func TestExtractAttributes_MergeWildcard(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("k8s.cluster.name", "cluster")
	attributes.PutStr("k8s.cluster.uid", "123")
	attributes.PutStr("k8s.pod.name", "pod")

	cs := config.ContextStatements{
		Attributes: []string{"k8s.cluster.*"},
	}

	_, values := ExtractAttributeKeys(attributes, cs)
	addDDTags(attributes, values)

	ddtags, ok := attributes.Get(ddtagsKey)
	assert.True(t, ok)

	assert.ElementsMatch(
		t,
		[]any{
			"k8s.cluster.name:cluster",
			"k8s.cluster.uid:123",
		},
		ddtags.Slice().AsRaw(),
	)

	// Merge should leave the source attributes untouched.
	clusterName, ok := attributes.Get("k8s.cluster.name")
	assert.True(t, ok)
	assert.Equal(t, "cluster", clusterName.AsString())

	clusterUID, ok := attributes.Get("k8s.cluster.uid")
	assert.True(t, ok)
	assert.Equal(t, "123", clusterUID.AsString())

	podName, ok := attributes.Get("k8s.pod.name")
	assert.True(t, ok)
	assert.Equal(t, "pod", podName.AsString())
}

func TestExtractAttributes_MoveWildcard(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("k8s.cluster.name", "cluster")
	attributes.PutStr("k8s.cluster.uid", "123")
	attributes.PutStr("k8s.pod.name", "pod")

	cs := config.ContextStatements{
		Attributes: []string{"k8s.cluster.*"},
	}

	keys, values := ExtractAttributeKeys(attributes, cs)
	addDDTags(attributes, values)

	for _, key := range keys {
		attributes.Remove(key)
	}

	ddtags, ok := attributes.Get(ddtagsKey)
	assert.True(t, ok)

	assert.ElementsMatch(
		t,
		[]any{
			"k8s.cluster.name:cluster",
			"k8s.cluster.uid:123",
		},
		ddtags.Slice().AsRaw(),
	)

	_, exists := attributes.Get("k8s.cluster.name")
	assert.False(t, exists)

	_, exists = attributes.Get("k8s.cluster.uid")
	assert.False(t, exists)

	// Non-selected attributes should remain.
	value, ok := attributes.Get("k8s.pod.name")
	assert.True(t, ok)
	assert.Equal(t, "pod", value.AsString())
}

func TestExtractAttributes_PreservesExistingDDTags(t *testing.T) {
	attributes := pcommon.NewMap()

	ddtags := attributes.PutEmptySlice(ddtagsKey)
	ddtags.AppendEmpty().SetStr("existing:value")

	attributes.PutStr("service.name", "my-service")

	cs := config.ContextStatements{
		Attributes: []string{"service.name"},
	}

	_, values := ExtractAttributeKeys(attributes, cs)
	addDDTags(attributes, values)

	result, _ := attributes.Get(ddtagsKey)

	assert.ElementsMatch(
		t,
		[]any{
			"existing:value",
			"service.name:my-service",
		},
		result.Slice().AsRaw(),
	)
}

func TestExtractAttributes_DifferentSourceAndDestination(t *testing.T) {
	source := pcommon.NewMap()
	source.PutStr("k8s.cluster.name", "cluster-a")
	source.PutStr("k8s.cluster.uid", "123")

	destination := pcommon.NewMap()

	cs := config.ContextStatements{
		Context:    config.Resource,
		Attributes: []string{"k8s.cluster.*"},
	}

	keys, values := ExtractAttributeKeys(source, cs)
	addDDTags(destination, values)

	for _, key := range keys {
		source.Remove(key)
	}

	ddtags, ok := destination.Get(ddtagsKey)
	assert.True(t, ok)
	assert.ElementsMatch(
		t,
		[]any{"k8s.cluster.name:cluster-a", "k8s.cluster.uid:123"},
		ddtags.Slice().AsRaw(),
	)

	// The source has no ddtags key of its own, and the moved keys are gone.
	_, ok = source.Get(ddtagsKey)
	assert.False(t, ok)
	_, ok = source.Get("k8s.cluster.name")
	assert.False(t, ok)
	_, ok = source.Get("k8s.cluster.uid")
	assert.False(t, ok)
}

func TestProcessRecordAttributes(t *testing.T) {
	tests := []struct {
		name                  string
		attributes            map[string]string
		cs                    config.ContextStatements
		resourceAttributeTags []string
		expectedDDTags        []string
		expectedAttributes    map[string]string
	}{
		{
			name: "resource",
			attributes: map[string]string{
				"team": "order",
			},
			cs: config.ContextStatements{
				Context: config.Resource,
				Mode:    config.Merge,
			},
			resourceAttributeTags: []string{
				"service.name:my-service",
				"k8s.namespace.name:my-namespace",
			},
			expectedDDTags: []string{
				"service.name:my-service",
				"k8s.namespace.name:my-namespace",
			},
			expectedAttributes: map[string]string{
				"team": "order",
			},
		},
		{
			name: "merge",
			attributes: map[string]string{
				"team":    "order",
				"message": "hello",
			},
			cs: config.ContextStatements{
				Context: config.Log,
				Mode:    config.Merge,
				Attributes: []string{
					"team",
				},
			},
			expectedDDTags: []string{
				"team:order",
			},
			expectedAttributes: map[string]string{
				"team":    "order",
				"message": "hello",
			},
		},
		{
			name: "move",
			attributes: map[string]string{
				"team":    "order",
				"message": "hello",
			},
			cs: config.ContextStatements{
				Context: config.Log,
				Mode:    config.Move,
				Attributes: []string{
					"team",
				},
			},
			expectedDDTags: []string{
				"team:order",
			},
			expectedAttributes: map[string]string{
				"message": "hello",
			},
		},
		{
			name: "no matching attributes",
			attributes: map[string]string{
				"team": "order",
			},
			cs: config.ContextStatements{
				Context: config.Log,
				Mode:    config.Merge,
				Attributes: []string{
					"service.name",
				},
			},
			expectedDDTags: nil,
			expectedAttributes: map[string]string{
				"team": "order",
			},
		},
		{
			name: "resource ignores record attributes",
			attributes: map[string]string{
				"team": "record-team",
			},
			cs: config.ContextStatements{
				Context: config.Resource,
				Mode:    config.Move,
				Attributes: []string{
					"team",
				},
			},
			resourceAttributeTags: []string{
				"team:resource-team",
			},
			expectedDDTags: []string{
				"team:resource-team",
			},
			expectedAttributes: map[string]string{
				"team": "record-team",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attributes := pcommon.NewMap()

			for key, value := range tt.attributes {
				attributes.PutStr(key, value)
			}

			ProcessRecordAttributes(
				attributes,
				tt.cs,
				tt.resourceAttributeTags,
			)

			ddtags, ok := attributes.Get("ddtags")

			if len(tt.expectedDDTags) == 0 {
				require.False(t, ok)
			} else {
				require.True(t, ok)

				slice := ddtags.Slice()
				require.Equal(t, len(tt.expectedDDTags), slice.Len())

				for i, expected := range tt.expectedDDTags {
					require.Equal(t, expected, slice.At(i).Str())
				}
			}

			for key, expected := range tt.expectedAttributes {
				value, ok := attributes.Get(key)
				require.True(t, ok, "expected attribute %q to exist", key)
				require.Equal(t, expected, value.Str())
			}

			expectedCount := len(tt.expectedAttributes)
			if len(tt.expectedDDTags) > 0 {
				expectedCount++
			}

			require.Equal(t, expectedCount, attributes.Len())
		})
	}
}
