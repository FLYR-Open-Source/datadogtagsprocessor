package extraction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getChild(t *testing.T, n *node, name string) *node {
	t.Helper()

	child, ok := n.children[name]
	assert.True(t, ok, "expected child %q to exist", name)

	return child
}

func TestLookup(t *testing.T) {
	tests := []struct {
		name       string
		attributes []string
		lookup     string
		expected   []string
	}{
		{
			name: "Should return all the k8s.cluster.* attributes",
			attributes: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
				"k8s.container.name",
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.node.name",
				"k8s.pod.name",
			},
			lookup:   "k8s.cluster",
			expected: []string{"k8s.cluster.name", "k8s.cluster.uid"},
		},
		{
			name: "Should return empty list",
			attributes: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
				"k8s.container.name",
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.node.name",
				"k8s.pod.name",
			},
			lookup:   "",
			expected: []string{},
		},
		{
			name: "Should return all k8s.* attributes",
			attributes: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
				"k8s.container.name",
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.node.name",
				"k8s.pod.name",
			},
			lookup: "k8s",
			expected: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
				"k8s.container.name",
				"k8s.deployment.name",
				"k8s.namespace.name",
				"k8s.node.name",
				"k8s.pod.name",
			},
		},
		{
			name: "Should return exact leaf",
			attributes: []string{
				"k8s.cluster.name",
				"k8s.cluster.uid",
			},
			lookup:   "k8s.cluster.name",
			expected: []string{"k8s.cluster.name"},
		},
		{
			name: "Should return empty for unknown namespace",
			attributes: []string{
				"k8s.cluster.name",
			},
			lookup:   "k8s.service",
			expected: []string{},
		},
		{
			name: "Should return empty for unknown root",
			attributes: []string{
				"k8s.cluster.name",
			},
			lookup:   "host",
			expected: []string{},
		},
		{
			name: "Should support single component key",
			attributes: []string{
				"service",
			},
			lookup:   "service",
			expected: []string{"service"},
		},
		{
			name:       "Should return empty from empty trie",
			attributes: []string{},
			lookup:     "k8s",
			expected:   []string{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := newNode()

			for _, attr := range test.attributes {
				root.insert(attr)
			}

			res := root.lookup(test.lookup)
			assert.ElementsMatch(t, test.expected, res)
		})
	}
}

func TestInsert(t *testing.T) {
	root := newNode()

	root.insert("k8s.cluster.name")

	k8s := getChild(t, root, "k8s")
	cluster := getChild(t, k8s, "cluster")
	name := getChild(t, cluster, "name")

	assert.False(t, k8s.isLeaf)
	assert.False(t, cluster.isLeaf)
	assert.True(t, name.isLeaf)
}

func TestInsert_SharedPrefix(t *testing.T) {
	root := newNode()

	root.insert("k8s.cluster.name")
	root.insert("k8s.cluster.uid")

	cluster := getChild(t, getChild(t, root, "k8s"), "cluster")

	name := getChild(t, cluster, "name")
	uid := getChild(t, cluster, "uid")

	assert.True(t, name.isLeaf)
	assert.True(t, uid.isLeaf)

	assert.ElementsMatch(t,
		[]string{
			"k8s.cluster.name",
			"k8s.cluster.uid",
		},
		cluster.keys,
	)
}
