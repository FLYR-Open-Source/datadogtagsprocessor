package extraction

import (
	"testing"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

func TestEnsureDDTags(t *testing.T) {
	t.Run("should create ddtags when it does not exist", func(t *testing.T) {
		attributes := pcommon.NewMap()

		ensureDDTags(attributes)

		ddtags, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Empty(t, ddtags.Slice().AsRaw())
	})

	t.Run("should not overwrite existing ddtags", func(t *testing.T) {
		attributes := pcommon.NewMap()

		ddtags := attributes.PutEmptySlice(ddtagsKey)
		ddtags.AppendEmpty().SetStr("existing:value")

		ensureDDTags(attributes)

		value, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Equal(t, []any{"existing:value"}, value.Slice().AsRaw())
	})
}

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

func TestAddDDTag(t *testing.T) {
	t.Run("should add attribute as ddtags", func(t *testing.T) {
		attributes := pcommon.NewMap()
		attributes.PutEmptySlice(ddtagsKey)
		attributes.PutStr("service.name", "my-service")

		addDDTag(attributes, "service.name")

		ddtags, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Equal(
			t,
			[]any{"service.name:my-service"},
			ddtags.Slice().AsRaw(),
		)
	})

	t.Run("should do nothing when attribute does not exist", func(t *testing.T) {
		attributes := pcommon.NewMap()
		attributes.PutEmptySlice(ddtagsKey)

		addDDTag(attributes, "service.name")

		ddtags, ok := attributes.Get(ddtagsKey)

		assert.True(t, ok)
		assert.Empty(t, ddtags.Slice().AsRaw())
	})
}

func TestExtractAttributes_Merge(t *testing.T) {
	attributes := pcommon.NewMap()
	attributes.PutStr("service.name", "my-service")
	attributes.PutStr("service.version", "1.2.3")

	cs := config.ContextStatements{
		Attributes: []string{"service.name"},
	}

	extractAttributes(attributes, cs, config.Merge)

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

	extractAttributes(attributes, cs, config.Move)

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

	extractAttributes(attributes, cs, config.Merge)

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

	extractAttributes(attributes, cs, config.Move)

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

	extractAttributes(attributes, cs, config.Merge)

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
