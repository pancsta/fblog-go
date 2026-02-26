package fblog

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

// HandlebarsTemplate wraps the Go text/template to mimic the registry behavior.
type HandlebarsTemplate struct {
	tpl *template.Template
}

// FblogHandlebarRegistry initializes the standard text/template with utility functions.
func FblogHandlebarRegistry(mainLineFormat, additionalValueFormat string) (*HandlebarsTemplate, error) {
	funcMap := template.FuncMap{
		"bold":        boldFunc,
		"cyan":        cyanFunc,
		"yellow":      yellowFunc,
		"red":         redFunc,
		"blue":        blueFunc,
		"purple":      purpleFunc,
		"green":       greenFunc,
		"color_rgb":   colorRGBFunc,
		"uppercase":   uppercaseFunc,
		"level_style": levelStyleFunc,
		"fixed_size":  fixedSizeFunc,
		"min_size":    minSizeFunc,
	}

	tpl := template.New("fblog").Funcs(funcMap)

	// Since we use the raw strings from Config, we need to parse them
	// Note: We use the text/template syntax in the Go config instead of handlebars syntax
	tpl, err := tpl.Parse(fmt.Sprintf(`{{define "main_line"}}%s{{end}}`, mainLineFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to parse main_line template: %w", err)
	}

	tpl, err = tpl.Parse(fmt.Sprintf(`{{define "additional_value"}}%s{{end}}`, additionalValueFormat))
	if err != nil {
		return nil, fmt.Errorf("failed to parse additional_value template: %w", err)
	}

	return &HandlebarsTemplate{tpl: tpl}, nil
}

// Render executes a named template and returns the resulting string.
func (h *HandlebarsTemplate) Render(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := h.tpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	// Return string without trailing newline directly (handled by logger)
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// Template Functions

func boldFunc(text string) string {
	return color.New(color.Bold).Sprint(text)
}

func cyanFunc(text string) string {
	return color.New(color.FgCyan).Sprint(text)
}

func yellowFunc(text string) string {
	return color.New(color.FgYellow).Sprint(text)
}

func redFunc(text string) string {
	return color.New(color.FgRed).Sprint(text)
}

func blueFunc(text string) string {
	return color.New(color.FgBlue).Sprint(text)
}

func purpleFunc(text string) string {
	return color.New(color.FgMagenta).Sprint(text)
}

func greenFunc(text string) string {
	return color.New(color.FgGreen).Sprint(text)
}

func colorRGBFunc(r, g, b int, text string) string {
	// fatih/color doesn't have a direct RGB builder that easily matches yansi,
	// but true color (24-bit) can be printed using RGB escape sequences:
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, text)
}

func uppercaseFunc(text string) string {
	return strings.ToUpper(text)
}

func levelStyleFunc(levelStr string) string {
	lvl := strings.ToLower(strings.TrimSpace(levelStr))
	var c *color.Color
	switch lvl {
	case "trace":
		c = color.New(color.FgCyan, color.Bold)
	case "debug":
		c = color.New(color.FgBlue, color.Bold)
	case "info":
		c = color.New(color.FgGreen, color.Bold)
	case "warn", "warning":
		c = color.New(color.FgYellow, color.Bold)
	case "error", "err", "fatal":
		c = color.New(color.FgRed, color.Bold)
	default:
		c = color.New(color.FgMagenta, color.Bold)
	}
	return c.Sprint(levelStr)
}

func fixedSizeFunc(size int, text string) string {
	if len(text) > size {
		// Truncate
		return text[:size]
	} else if len(text) < size {
		// Pad with spaces on the left
		return strings.Repeat(" ", size-len(text)) + text
	}
	return text
}

func minSizeFunc(size int, text string) string {
	if len(text) < size {
		// Pad with spaces on the left
		return strings.Repeat(" ", size-len(text)) + text
	}
	return text
}
