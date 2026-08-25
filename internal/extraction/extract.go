package extraction

import (
	"slices"
	"strings"

	"github.com/FLYR-Open-Source/datadogtagsprocessor/internal/config"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	// ddtagsKey is the attribute that holds the Datadog tags.
	ddtagsKey = "ddtags"
	// tagSeparator joins the tags inside the ddtags attribute. Datadog
	// splits the attribute on it to get the individual tags.
	tagSeparator = ","
)

// ResourceAttributes is the attribute map of a resource.
type ResourceAttributes = pcommon.Map

// Attributes is the attribute map of a single record.
type Attributes = pcommon.Map

// buildAttributeLookup groups the attribute keys that fall under each
// wildcard selection.
//
// The result is keyed by the wildcard namespace and the keys keep the
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

// getDDTags returns the tags already held by the ddtags attribute.
//
// Datadog reads ddtags as a single string of comma separated tags, so the
// value is returned as it will be sent. A missing attribute gives an empty
// string.
func getDDTags(attributes pcommon.Map) string {
	value, ok := attributes.Get(ddtagsKey)
	if !ok {
		return ""
	}

	if value.Type() != pcommon.ValueTypeSlice {
		return value.AsString()
	}

	slice := value.Slice()
	tags := make([]string, 0, slice.Len())

	for i := 0; i < slice.Len(); i++ {
		tags = append(tags, slice.At(i).AsString())
	}

	return strings.Join(tags, tagSeparator)
}

// getTagsFormatted formats the given keys as Datadog tags.
//
// Every key that exists in the attributes becomes a "key:value" entry.
// Keys that do not exist are skipped.
func getTagsFormatted(attributes pcommon.Map, keys []string) []string {
	values := make([]string, 0, len(keys))

	for _, key := range keys {
		value, ok := attributes.Get(key)
		if ok {
			values = append(values, key+":"+value.AsString())
		}
	}

	return values
}

// addDDTags appends the given values to the ddtags attribute.
//
// The tags are written as one comma separated string, the format Datadog
// expects. Tags already in the attribute are kept ahead of the new ones.
func addDDTags(attributes Attributes, values []string) {
	if len(values) == 0 {
		return
	}

	tags := strings.Join(values, tagSeparator)

	if existing := getDDTags(attributes); existing != "" {
		tags = existing + tagSeparator + tags
	}

	attributes.PutStr(ddtagsKey, tags)
}

// ExtractAttributeKeys resolves the statement's selections against the
// given attributes.
//
// It returns the matched attribute keys in selection order and the same
// keys formatted as "key:value" tags.
func ExtractAttributeKeys(attributes pcommon.Map, cs config.CompiledStatement) (attributeKeys, ddTagsFormat []string) {
	var lookup map[string][]string
	if cs.HasWildcards {
		lookup = buildAttributeLookup(attributes, cs.Attributes)
	}

	// At least one key per selection; wildcards may add more.
	attributeKeys = make([]string, 0, len(cs.Attributes))

	for _, selected := range cs.Attributes {
		if selected.Wildcard {
			attributeKeys = append(attributeKeys, lookup[selected.Key]...)
			continue
		}

		if _, ok := attributes.Get(selected.Key); ok {
			attributeKeys = append(attributeKeys, selected.Key)
		}
	}

	ddTagsFormat = getTagsFormatted(attributes, attributeKeys)
	return attributeKeys, ddTagsFormat
}

// RemoveAttributes deletes the given keys from the attributes.
//
// The keys are removed in a single pass over the map, unlike per-key Remove
// calls which each rescan it. It is a no-op when keys is empty.
func RemoveAttributes(attributes pcommon.Map, keys []string) {
	if len(keys) == 0 {
		return
	}

	attributes.RemoveIf(func(key string, _ pcommon.Value) bool {
		return slices.Contains(keys, key)
	})
}

// ProcessRecordAttributes applies a statement to the attributes of a single
// record.
//
// For resource statements it only appends the precomputed resource tags.
// For record statements it resolves the selections, appends the tags and,
// in move mode, removes the matched attributes.
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
		RemoveAttributes(attributes, attributeKeys)
	}
}
