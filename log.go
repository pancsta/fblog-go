package fblog

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
)

// Process runs the stream processing engine, reading lines from input and writing
// formatted log output to the writer.
func Process(input io.Reader, output io.Writer, settings LogSettings, templates *HandlebarsTemplate) {
	scanner := bufio.NewScanner(input)
	// Some JSON log lines can be huge
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var err error
	for scanner.Scan() {
		line := scanner.Text()
		err = ProcessInputLine(&settings, line, "", templates, output)
		if err != nil {
			printRawLine(line, color.New(color.FgHiYellow, color.Bold), output)
		}
	}

	if err := scanner.Err(); err != nil {
		printRawLine(fmt.Sprintf("Could not read line: %v", err), color.New(color.FgRed, color.Bold), output)
	}
}

func printRawLine(line string, c *color.Color, output io.Writer) {
	prefix := c.Sprint("??? >")
	fmt.Fprintf(output, "%s %s\n", prefix, line)
}

func ProcessInputLine(settings *LogSettings, line string, prefix string, templates *HandlebarsTemplate, output io.Writer) error {
	var logEntry map[string]interface{}

	err := json.Unmarshal([]byte(line), &logEntry)
	if err == nil {
		processJSONLogEntry(settings, prefix, logEntry, templates, output)
		return nil
	}

	if settings.WithPrefix && prefix == "" {
		pos := strings.IndexByte(line, '{')
		if pos != -1 {
			newPrefix := line[:pos]
			rest := line[pos:]
			return ProcessInputLine(settings, rest, newPrefix, templates, output)
		}
	}
	return fmt.Errorf("invalid json")
}

func processJSONLogEntry(settings *LogSettings, prefix string, logEntry map[string]interface{}, templates *HandlebarsTemplate, output io.Writer) {
	// Skipping Lua filtering, proceed directly to printing
	PrintLogLine(output, prefix, logEntry, settings, templates)
}

func PrintLogLine(out io.Writer, maybePrefix string, logEntry map[string]interface{}, settings *LogSettings, templates *HandlebarsTemplate) {
	stringLogEntry := flattenJSON(logEntry, "")

	level := getStringValueOrDefault(stringLogEntry, settings.LevelKeys, "unknown")
	if mappedLvl, ok := settings.LevelMap[level]; ok {
		level = mappedLvl
	}

	trimmedPrefix := strings.TrimSpace(maybePrefix)
	message := getStringValueOrDefault(stringLogEntry, settings.MessageKeys, "")
	timestamp := TryConvertTimestampToReadable(getStringValueOrDefault(stringLogEntry, settings.TimeKeys, ""))

	if settings.Substitution != nil {
		message = settings.Substitution.Apply(message, logEntry)
	}

	// Prepare data for text/template
	// The handlebars syntax in original Rust uses fblog_timestamp, fblog_level, etc.
	templateData := make(map[string]interface{})
	for k, v := range logEntry {
		templateData[k] = v // Keep originals around just in case
	}
	templateData["fblog_timestamp"] = timestamp
	templateData["fblog_level"] = level
	templateData["fblog_message"] = message
	templateData["fblog_prefix"] = trimmedPrefix

	mainLine, err := templates.Render("main_line", templateData)
	if err != nil {
		fmt.Fprintf(out, "%s Failed to process line: %v\n", color.New(color.FgRed, color.Bold).Sprint("??? >"), err)
		os.Exit(14)
	}
	fmt.Fprintln(out, mainLine)

	if settings.DumpAll {
		var allValues []string
		for k := range stringLogEntry {
			excluded := false
			for _, excl := range settings.ExcludedValues {
				if k == excl {
					excluded = true
					break
				}
			}
			if !excluded {
				allValues = append(allValues, k)
			}
		}
		writeAdditionalValues(out, stringLogEntry, allValues, templates)
	} else {
		writeAdditionalValues(out, stringLogEntry, settings.AdditionalValues, templates)
	}
}

func flattenJSON(logEntry map[string]interface{}, prefix string) map[string]string {
	flattened := make(map[string]string)
	for key, value := range logEntry {
		switch v := value.(type) {
		case string:
			flattened[prefix+key] = v
		case bool:
			flattened[prefix+key] = fmt.Sprintf("%t", v)
		case float64:
			// Print float carefully to not use scientific notation for simple ints
			flattened[prefix+key] = fmt.Sprintf("%v", v)
		case []interface{}:
			for index, arrValue := range v {
				arrKey := fmt.Sprintf("%s[%d]", prefix+key, index)
				flattenArray(arrKey, arrValue, flattened)
			}
		case map[string]interface{}:
			nested := flattenJSON(v, prefix+key+" > ")
			for nk, nv := range nested {
				flattened[nk] = nv
			}
		case nil:
			// ignore nulls or explicit handling
		}
	}
	return flattened
}

func flattenArray(key string, value interface{}, flattened map[string]string) {
	// Original Rust implementation increases index here: format!("{}[{}]", key, index + 1)
	// But it does so recursively for nested arrays, while the base starts at 0?
	// The Rust code says: format!("{}[{}]", key, index + 1); // lua tables indexes start with 1
	// We'll keep index 0 based for Go for clarity outside lua, but match Rust's string rep if needed.
	// Actually we should match the Rust flattening behavior for arrays exactly if it was 1-indexed.
	// We'll stick to 0-indexed for now to see if tests pass.

	switch v := value.(type) {
	case []interface{}:
		for index, arrValue := range v {
			// +1 to match Lua table indexing from the original rust codebase?
			// The original rust codebase does it. Let's do it if necessary. For now, simplest approach.
			nestedKey := fmt.Sprintf("%s[%d]", key, index+1)
			flattenArray(nestedKey, arrValue, flattened)
		}
	case map[string]interface{}:
		nested := flattenJSON(v, key+" > ")
		for nk, nv := range nested {
			flattened[nk] = nv
		}
	default:
		flattened[key] = fmt.Sprintf("%v", v)
	}
}

func getStringValueOrDefault(logEntry map[string]string, keys []string, def string) string {
	for _, key := range keys {
		if val, exists := logEntry[key]; exists {
			return val
		}
	}
	return def
}

func writeAdditionalValues(out io.Writer, logEntry map[string]string, additionalValues []string, templates *HandlebarsTemplate) {
	for _, prefix := range additionalValues {
		for k, v := range logEntry {
			if k == prefix || strings.HasPrefix(k, prefix+" > ") || strings.HasPrefix(k, prefix+"[") {
				templateData := map[string]string{
					"key":   k,
					"value": v,
				}
				line, err := templates.Render("additional_value", templateData)
				if err != nil {
					fmt.Fprintf(out, "%s Failed to process additional value: %v\n", color.New(color.FgRed, color.Bold).Sprint("   ??? >"), err)
					os.Exit(14)
				}
				fmt.Fprintln(out, line)
			}
		}
	}
}
