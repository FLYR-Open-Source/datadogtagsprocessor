package extraction

import (
	"fmt"
	"strings"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/datadogtagsprocessor/internal/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	ddtagsKey = "ddtags"
)

func ensureDDTags(attributes pcommon.Map) {
	if _, ok := attributes.Get(ddtagsKey); !ok {
		attributes.PutEmptySlice(ddtagsKey)
	}
}

func hasWildcardSuffix(attribute string) (string, bool) {
	return strings.CutSuffix(attribute, ".*")
}

func wildcardNamespaces(attributes []string) []string {
	var namespaces []string

	for _, attr := range attributes {
		if namespace, ok := hasWildcardSuffix(attr); ok {
			namespaces = append(namespaces, namespace)
		}
	}

	return namespaces
}

func buildAttributeLookup(attributes pcommon.Map, namespaces []string) *node {
	root := newNode()

	for key := range attributes.AsRaw() {
		for _, namespace := range namespaces {
			if key == namespace || strings.HasPrefix(key, namespace+".") {
				root.insert(key)
				break
			}
		}
	}

	return root
}

func lookupSelectedAttributes(root *node, attributes pcommon.Map, selected string) []string {
	namespace, wildcard := hasWildcardSuffix(selected)

	if wildcard {
		return root.lookup(namespace)
	}

	if _, ok := attributes.Get(namespace); ok {
		return []string{namespace}
	}

	return nil
}

func addDDTag(attributes pcommon.Map, key string) {
	value, ok := attributes.Get(key)
	if !ok {
		return
	}

	ddtags, _ := attributes.Get(ddtagsKey)

	ddtags.Slice().AppendEmpty().SetStr(
		fmt.Sprintf("%s:%s", key, value.AsString()),
	)
}

func extractAttributes(attributes pcommon.Map, cs config.ContextStatements, mode config.Mode) {
	ensureDDTags(attributes)

	var lookup *node
	if namespaces := wildcardNamespaces(cs.Attributes); len(namespaces) > 0 {
		lookup = buildAttributeLookup(attributes, namespaces)
	}

	for _, selectedAttribute := range cs.Attributes {
		keys := lookupSelectedAttributes(lookup, attributes, selectedAttribute)

		for _, key := range keys {
			addDDTag(attributes, key)

			if mode == config.Move {
				attributes.Remove(key)
			}
		}
	}
}
