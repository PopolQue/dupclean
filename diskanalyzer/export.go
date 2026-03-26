package diskanalyzer

import (
	"encoding/json"
	"io"
)

// ExportJSON writes the AnalysisResult as JSON to the given writer.
// The DirNode.Parent field is excluded from JSON serialization to avoid cycles.
func ExportJSON(result *AnalysisResult, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// ExportJSONPretty writes the AnalysisResult as indented JSON.
func ExportJSONPretty(result *AnalysisResult, w io.Writer) error {
	return ExportJSON(result, w)
}

// ExportJSONCompact writes the AnalysisResult as compact JSON (no indentation).
func ExportJSONCompact(result *AnalysisResult, w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(result)
}
