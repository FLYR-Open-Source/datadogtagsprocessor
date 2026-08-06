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

// buildAttributeLookup groups the attribute keys that fall under each
// wildcard selection's namespace, keyed by that namespace. Keys keep the
// attribute map's insertion order. A key matching several namespaces is
// grouped under the first match only.
func buildAttributeLookup(attributes pcommon.Map, selections []config.CompiledAttribute) map[string][]string {
	lookup := make(map[string][]string, len(selections))

	attributes.Range(func(k string, _ pcommon.Value) bool {
		for _, selected := range selections {
			if !selected.Wildcard {
				continue
			}

			if k == selected.Key || strings.HasPrefix(k, selected.Prefix) {
				lookup[selected.Key] = append(lookup[selected.Key], k)
				return true
			}
		}
		return true
	})

	return lookup
}

func lookupSelectedAttributes(lookup map[string][]string, attributes pcommon.Map, selected config.CompiledAttribute) []string {
	if selected.Wildcard {
		return lookup[selected.Key]
	}

	if _, ok := attributes.Get(selected.Key); ok {
		return []string{selected.Key}
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

func ExtractAttributeKeys(attributes pcommon.Map, cs config.CompiledStatement) (attributeKeys, ddTagsFormat []string) {
	var lookup map[string][]string
	if cs.HasWildcards {
		lookup = buildAttributeLookup(attributes, cs.Attributes)
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
	cs config.CompiledStatement,
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
