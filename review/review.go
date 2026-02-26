package main

import (
	"os"

	"github.com/pancsta/fblog-go"
)

func main() {
	config := fblog.NewDefaultConfig()
	template, _ := fblog.FblogHandlebarRegistry(config.MainLineFormat, config.AdditionalValueFormat)
	logSettings := fblog.NewDefaultLogSettings()
	logSettings.DumpAll = true

	logEntry := map[string]interface{}{
		"message": "something happened",
		"time":    "2017-07-06T15:21:16",
		"process": "rust",
		"fu":      "bower",
		"level":   "info",
	}

	fblog.PrintLogLine(os.Stdout, "", logEntry, &logSettings, template)
}
