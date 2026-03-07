package ingest

import (
	"strconv"
	"strings"
)

const (
	categoryBullet    int16 = 1
	categoryBlitz     int16 = 2
	categoryRapid     int16 = 3
	categoryClassical int16 = 4
)

// parseTimeControlCategory derives a time control category from a raw time
// control string. Returns the category and false, or 0 and true if the
// string is empty, missing, or unparseable.
func parseTimeControlCategory(tc string) (int16, bool) {
	if tc == "" || tc == "-" || tc == "?" {
		return 0, true
	}

	// Multi-stage: "40/5400:1800" — sum all stage times.
	// Single stage: "300+5" — base+increment.
	totalSeconds := 0
	stages := strings.Split(tc, ":")

	for _, stage := range stages {
		// Strip move count prefix if present (e.g., "40/5400" → "5400").
		if idx := strings.IndexByte(stage, '/'); idx >= 0 {
			stage = stage[idx+1:]
		}

		// Parse base+increment.
		parts := strings.SplitN(stage, "+", 2)
		base, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, true
		}

		increment := 0
		if len(parts) == 2 {
			increment, err = strconv.Atoi(parts[1])
			if err != nil {
				return 0, true
			}
		}

		// Estimated game duration assumes ~40 moves (FIDE convention).
		totalSeconds += base + 40*increment
	}

	minutes := totalSeconds / 60

	switch {
	case minutes < 3:
		return categoryBullet, false
	case minutes < 10:
		return categoryBlitz, false
	case minutes < 60:
		return categoryRapid, false
	default:
		return categoryClassical, false
	}
}
