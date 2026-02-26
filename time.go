package fblog

import (
	"math"
	"strconv"
	"time"
)

// TryConvertTimestampToReadable tries to convert a given string representing a timestamp
// into a readable ISO8601 (RFC3339) string. If it cannot be parsed as a timestamp,
// it returns the input string as-is.
func TryConvertTimestampToReadable(input string) string {
	if input == "" {
		return input
	}

	timestamp, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return input
	}

	iso8601, ok := timestampToISO8601(timestamp)
	if ok {
		return iso8601
	}

	return input
}

func timestampToISO8601(timestamp int64) (string, bool) {
	now := time.Now()

	// Try interpreting as seconds
	dtSecs := time.Unix(timestamp, 0).UTC()
	// Try interpreting as milliseconds
	dtMillis := time.UnixMilli(timestamp).UTC()

	// Check which one is closer to now
	diffSecs := math.Abs(float64(now.Sub(dtSecs).Milliseconds()))
	diffMillis := math.Abs(float64(now.Sub(dtMillis).Milliseconds()))

	if diffSecs < diffMillis {
		return dtSecs.Format("2006-01-02T15:04:05.000Z"), true
	} else {
		return dtMillis.Format("2006-01-02T15:04:05.000Z"), true
	}
}
