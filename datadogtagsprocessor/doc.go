//go:generate make mdatagen

// Package datadogtagsprocessor contains the logic to move attributes and
// resource attributes to the `attributes["ddtags"]` array so these values
// can appear as labels in Trace and Log resources.
package datadogtagsprocessor
