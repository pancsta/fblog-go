package fblog

import (
	"strings"
	"testing"
)

func TestSubstitutionCommonFormats(t *testing.T) {
	testPlaceholderFormat(t, "{key}")
	testPlaceholderFormat(t, "[key]")
	testPlaceholderFormat(t, "%key%")
	testPlaceholderFormat(t, "${key}")
}

func testPlaceholderFormat(t *testing.T, placeholder string) {
	subst, err := NewSubstitution("", placeholder)
	if err != nil {
		t.Fatalf("Failed to create substitution for placeholder %s: %v", placeholder, err)
	}

	msg := strings.Replace("Tapping fingers as a way to {placeholder}", "{placeholder}", placeholder, 1)

	context := map[string]interface{}{
		"key": "speak",
	}

	logEntry := map[string]interface{}{
		subst.ContextKey: context,
	}

	result := subst.Apply(msg, logEntry)
	expected := "Tapping fingers as a way to \x1b[33;1mspeak\x1b[0m"

	if result != expected {
		t.Errorf("Failed to substitute with placeholder format %s: expected %q, got %q", placeholder, expected, result)
	}
}

func TestSubstitutionPlaceholderNotInContext(t *testing.T) {
	subst, _ := NewSubstitution("", "")
	msg := "substituted: {subst}, ignored: {ignored}"
	context := map[string]interface{}{
		"subst": "no brackets!",
	}
	logEntry := map[string]interface{}{
		subst.ContextKey: context,
	}

	result := subst.Apply(msg, logEntry)
	expected := "substituted: \x1b[33;1mno brackets!\x1b[0m, ignored: \x1b[2m{\x1b[0m\x1b[31;1mignored\x1b[0m\x1b[2m}\x1b[0m"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestSubstitutionArrayContext(t *testing.T) {
	subst, _ := NewSubstitution("", "")
	msg := "text: {0}, number: {1}, bool: {2}, ignored: {3}"
	context := []interface{}{
		"better than sleeping",
		9,
		true,
	}
	logEntry := map[string]interface{}{
		subst.ContextKey: context,
	}

	result := subst.Apply(msg, logEntry)
	// number will be formatted as int/float with cyan
	expected := "text: \x1b[33;1mbetter than sleeping\x1b[0m, number: \x1b[36;1m9\x1b[0m, bool: \x1b[32;1mtrue\x1b[0m, ignored: \x1b[2m{\x1b[0m\x1b[31;1m3\x1b[0m\x1b[2m}\x1b[0m"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
