package fblog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
)

// Substitution evaluates an input template against a context object and substitutes values.
type Substitution struct {
	ContextKey        string
	PlaceholderPrefix string
	PlaceholderSuffix string
	PlaceholderRegex  *regexp.Regexp
}

const (
	DefaultPlaceholderFormat = "{key}"
	KeyDelimiter             = "key"
	DefaultContextKey        = "context"
)

// NewSubstitution creates a new substitution struct
func NewSubstitution(contextKey, placeholderFormat string) (*Substitution, error) {
	if contextKey == "" {
		contextKey = DefaultContextKey
	}
	if placeholderFormat == "" {
		placeholderFormat = DefaultPlaceholderFormat
	}

	parts := strings.SplitN(placeholderFormat, KeyDelimiter, 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("the identifier `key` is missing in format %s", placeholderFormat)
	}

	prefix := parts[0]
	suffix := parts[1]

	// Create regex matching Prefix(WordBytes|Dash)Suffix
	// e.g. \{([a-zA-Z0-9_\-]+)\}
	pattern := fmt.Sprintf(`%s([a-zA-Z0-9_\-]+)%s`, regexp.QuoteMeta(prefix), regexp.QuoteMeta(suffix))
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("regular expression could not be created for format: %w", err)
	}

	return &Substitution{
		ContextKey:        contextKey,
		PlaceholderPrefix: prefix,
		PlaceholderSuffix: suffix,
		PlaceholderRegex:  regex,
	}, nil
}

// Apply searches the message for placeholders and substitutes them with values found in the context object.
func (s *Substitution) Apply(message string, logEntry map[string]interface{}) string {
	contextValue, ok := logEntry[s.ContextKey]
	if !ok || contextValue == nil {
		return message
	}

	return s.PlaceholderRegex.ReplaceAllStringFunc(message, func(match string) string {
		// Extract key from match group
		matches := s.PlaceholderRegex.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}
		key := matches[1]

		value := extractValue(contextValue, key)
		if value == nil {
			// Dim brackets, red bold key
			return color.New(color.Faint).Sprint(s.PlaceholderPrefix) + color.New(color.FgRed, color.Bold).Sprint(key) + color.New(color.Faint).Sprint(s.PlaceholderSuffix)
		}

		return colorFormat(value)
	})
}

func extractValue(context interface{}, key string) interface{} {
	switch v := context.(type) {
	case map[string]interface{}:
		return v[key]
	case []interface{}:
		index, err := strconv.Atoi(key)
		if err == nil && index >= 0 && index < len(v) {
			return v[index]
		}
	}
	return nil
}

func colorFormat(value interface{}) string {
	switch v := value.(type) {
	case string:
		return color.New(color.FgYellow, color.Bold).Sprint(v)
	case float64:
		// JSON numbers are parsed as float64 usually
		return color.New(color.FgCyan, color.Bold).Sprint(strconv.FormatFloat(v, 'f', -1, 64))
	case int:
		return color.New(color.FgCyan, color.Bold).Sprint(strconv.Itoa(v))
	case []interface{}:
		return colorFormatArray(v)
	case map[string]interface{}:
		return colorFormatMap(v)
	case bool:
		if v {
			return color.New(color.FgGreen, color.Bold).Sprint("true")
		}
		return color.New(color.FgRed, color.Bold).Sprint("false")
	case nil:
		return color.New(color.Bold).Sprint("null")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func colorFormatArray(a []interface{}) string {
	dim := color.New(color.Faint)
	var sb strings.Builder
	sb.WriteString(dim.Sprint("["))
	for i, val := range a {
		if i > 0 {
			sb.WriteString(dim.Sprint(", "))
		}
		sb.WriteString(colorFormat(val))
	}
	sb.WriteString(dim.Sprint("]"))
	return sb.String()
}

func colorFormatMap(m map[string]interface{}) string {
	dim := color.New(color.Faint)
	magenta := color.New(color.FgMagenta)

	var sb strings.Builder
	sb.WriteString(dim.Sprint("{"))
	i := 0
	for k, v := range m {
		if i > 0 {
			sb.WriteString(dim.Sprint(", "))
		}
		sb.WriteString(magenta.Sprint(k))
		sb.WriteString(dim.Sprint(": "))
		sb.WriteString(colorFormat(v))
		i++
	}
	sb.WriteString(dim.Sprint("}"))
	return sb.String()
}
