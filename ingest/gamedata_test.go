package ingest

import "testing"

func TestParseDate(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		wantYear, wantMonth    int16
		wantDay                int16
		yearNull, monthNull    bool
		dayNull                bool
	}{
		{"full date", "2024.01.15", 2024, 1, 15, false, false, false},
		{"unknown day", "2024.01.??", 2024, 1, 0, false, false, true},
		{"unknown month and day", "2024.??.??", 2024, 0, 0, false, true, true},
		{"all unknown", "????.??.??", 0, 0, 0, true, true, true},
		{"empty", "", 0, 0, 0, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			y, yNull, m, mNull, d, dNull := parseDate(tt.input)
			if yNull != tt.yearNull {
				t.Errorf("yearNull = %v, want %v", yNull, tt.yearNull)
			}
			if !yNull && y != tt.wantYear {
				t.Errorf("year = %d, want %d", y, tt.wantYear)
			}
			if mNull != tt.monthNull {
				t.Errorf("monthNull = %v, want %v", mNull, tt.monthNull)
			}
			if !mNull && m != tt.wantMonth {
				t.Errorf("month = %d, want %d", m, tt.wantMonth)
			}
			if dNull != tt.dayNull {
				t.Errorf("dayNull = %v, want %v", dNull, tt.dayNull)
			}
			if !dNull && d != tt.wantDay {
				t.Errorf("day = %d, want %d", d, tt.wantDay)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	tests := []struct {
		input string
		want  int16
	}{
		{"1-0", 1},
		{"0-1", -1},
		{"1/2-1/2", 0},
		{"*", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseResult(tt.input)
			if got != tt.want {
				t.Errorf("parseResult(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
