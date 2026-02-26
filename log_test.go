package fblog

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
)

func withoutStyle(styled string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(styled, "")
}

func fblogHandlebarRegistryDefaultFormat() (*HandlebarsTemplate, error) {
	config := NewDefaultConfig()
	return FblogHandlebarRegistry(config.MainLineFormat, config.AdditionalValueFormat)
}

func TestWriteLogEntry(t *testing.T) {
	templates, _ := fblogHandlebarRegistryDefaultFormat()
	logSettings := NewDefaultLogSettings()

	var out bytes.Buffer

	logEntry := map[string]interface{}{
		"message": "something happened",
		"time":    "2017-07-06T15:21:16",
		"process": "rust",
		"level":   "info",
	}

	PrintLogLine(&out, "", logEntry, &logSettings, templates)

	result := withoutStyle(out.String())
	expected := "2017-07-06T15:21:16  INFO: something happened\n"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestWriteLogEntryDumpAll(t *testing.T) {
	templates, _ := fblogHandlebarRegistryDefaultFormat()
	logSettings := NewDefaultLogSettings()
	logSettings.DumpAll = true // exclude nothing

	var out bytes.Buffer

	logEntry := map[string]interface{}{
		"message": "something happened",
		"time":    "2017-07-06T15:21:16",
		"process": "rust",
		"fu":      "bower",
		"level":   "info",
	}

	PrintLogLine(&out, "", logEntry, &logSettings, templates)

	result := withoutStyle(out.String())

	// Because Go maps are unordered, the output order of DumpAll is technically indeterminate
	// unless we sort keys in `log.go`. The Rust version used BTreeMap or IndexMap for stable ordering.
	// For testing, we just check if it contains the keys and values.

	if !strings.Contains(result, "2017-07-06T15:21:16  INFO: something happened") {
		t.Errorf("missing main line: %q", result)
	}
	if !strings.Contains(result, "process: rust") {
		t.Errorf("missing process field: %q", result)
	}
	if !strings.Contains(result, "fu: bower") {
		t.Errorf("missing fu field: %q", result)
	}
}

func TestWriteLogEntryWithArray(t *testing.T) {
	templates, _ := fblogHandlebarRegistryDefaultFormat()
	logSettings := NewDefaultLogSettings()
	logSettings.AddAdditionalValues([]string{"process", "fu"})

	var out bytes.Buffer

	logEntry := map[string]interface{}{
		"message": "something happened",
		"time":    "2017-07-06T15:21:16",
		"process": "rust",
		"fu":      []interface{}{"bower"},
		"level":   "info",
	}

	PrintLogLine(&out, "", logEntry, &logSettings, templates)

	result := withoutStyle(out.String())

	expectedMain := "2017-07-06T15:21:16  INFO: something happened"
	if !strings.Contains(result, expectedMain) {
		t.Errorf("missing main line: %q", result)
	}
	if !strings.Contains(result, "process: rust") {
		t.Errorf("missing process field: %q", result)
	}
	if !strings.Contains(result, `fu[0]: bower`) {
		t.Errorf("missing or incorrect fu[0] field: %q", result)
	}
}
