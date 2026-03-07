package ingest

import (
	"strconv"
	"strings"
)

// parseDate parses a PGN date string (YYYY.MM.DD) into separate nullable
// components. Unknown parts use "??" in PGN.
func parseDate(s string) (year int16, yearNull bool, month int16, monthNull bool, day int16, dayNull bool) {
	if s == "" {
		return 0, true, 0, true, 0, true
	}

	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return 0, true, 0, true, 0, true
	}

	if y, err := strconv.Atoi(parts[0]); err == nil {
		year = int16(y)
	} else {
		yearNull = true
	}

	if m, err := strconv.Atoi(parts[1]); err == nil {
		month = int16(m)
	} else {
		monthNull = true
	}

	if d, err := strconv.Atoi(parts[2]); err == nil {
		day = int16(d)
	} else {
		dayNull = true
	}

	return
}

// parseResult maps a PGN result string to a numeric value.
// 1-0 → 1 (white win), 0-1 → -1 (black win), 1/2-1/2 → 0 (draw), * → 0.
func parseResult(s string) int16 {
	switch s {
	case "1-0":
		return 1
	case "0-1":
		return -1
	default:
		return 0
	}
}
