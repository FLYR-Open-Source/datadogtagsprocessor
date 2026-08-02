package extraction

import (
	"strings"
)

type node struct {
	children map[string]*node
	keys     []string
	isLeaf   bool
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
	}
}

func (n *node) insert(key string) {
	current := n
	current.keys = append(current.keys, key)

	parts := strings.SplitSeq(key, ".")
	for part := range parts {
		child, ok := current.children[part]
		if !ok {
			child = newNode()
			current.children[part] = child
		}

		current = child
		current.keys = append(current.keys, key)
	}

	current.isLeaf = true
}

func (n *node) lookup(namespace string) []string {
	if namespace == "" {
		return []string{}
	}

	current := n

	parts := strings.SplitSeq(namespace, ".")
	for part := range parts {
		child, ok := current.children[part]
		if !ok {
			return []string{}
		}
		current = child
	}

	return current.keys
}
