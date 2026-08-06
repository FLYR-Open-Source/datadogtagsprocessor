package extraction

import (
	"strings"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	ddtagsKey = "ddtags"
)

type ResourceAttributes = pcommon.Map
type Attributes = pcommon.Map

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

// buildAttributeLookup groups the attribute keys that fall under each
// wildcard namespace, keyed by that namespace. Keys keep the attribute map's
// insertion order. A key matching several namespaces is grouped under the
// first match only.
func buildAttributeLookup(attributes pcommon.Map, namespaces []string) map[string][]string {
	lookup := make(map[string][]string, len(namespaces))

	prefixes := make([]string, len(namespaces))
	for i, namespace := range namespaces {
		prefixes[i] = namespace + "."
	}

	attributes.Range(func(k string, _ pcommon.Value) bool {
		for i, namespace := range namespaces {
			if k == namespace || strings.HasPrefix(k, prefixes[i]) {
				lookup[namespace] = append(lookup[namespace], k)
				return true
			}
		}
		return true
	})

	return lookup
}

func lookupSelectedAttributes(lookup map[string][]string, attributes pcommon.Map, selected string) []string {
	namespace, wildcard := hasWildcardSuffix(selected)

	if wildcard {
		return lookup[namespace]
	}

	if _, ok := attributes.Get(namespace); ok {
		return []string{namespace}
	}

	return nil
}

func getDDTags(attributes pcommon.Map) pcommon.Slice {
	value, ok := attributes.Get(ddtagsKey)
	if ok {
		return value.Slice()
	}

	return attributes.PutEmptySlice(ddtagsKey)
}

func getTagsFormatted(attributes pcommon.Map, keys []string) []string {
	var values []string

	for _, key := range keys {
		value, ok := attributes.Get(key)
		if ok {
			values = append(values, key+":"+value.AsString())
		}
	}

	return values
}

func addDDTags(attributes Attributes, values []string) {
	ddtags := getDDTags(attributes)

	for _, value := range values {
		ddtags.AppendEmpty().SetStr(value)
	}
}

func ExtractAttributeKeys(attributes pcommon.Map, cs config.ContextStatements) (attributeKeys, ddTagsFormat []string) {
	var lookup map[string][]string
	if namespaces := wildcardNamespaces(cs.Attributes); len(namespaces) > 0 {
		lookup = buildAttributeLookup(attributes, namespaces)
	}

	for _, selectedAttribute := range cs.Attributes {
		keys := lookupSelectedAttributes(lookup, attributes, selectedAttribute)
		attributeKeys = append(attributeKeys, keys...)
	}

	ddTagsFormat = getTagsFormatted(attributes, attributeKeys)
	return attributeKeys, ddTagsFormat
}

func ProcessRecordAttributes(
	attributes pcommon.Map,
	cs config.ContextStatements,
	resourceAttributeKeyValues []string,
) {
	if cs.Context == config.Resource {
		if len(resourceAttributeKeyValues) > 0 {
			addDDTags(attributes, resourceAttributeKeyValues)
		}
		return
	}

	attributeKeys, attributeKeyValues := ExtractAttributeKeys(attributes, cs)

	if len(attributeKeyValues) > 0 {
		addDDTags(attributes, attributeKeyValues)
	}

	if cs.Mode == config.Move {
		for _, key := range attributeKeys {
			attributes.Remove(key)
		}
	}
}
