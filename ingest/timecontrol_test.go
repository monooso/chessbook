package ingest

import "testing"

func TestParseTimeControlCategory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantCat  int16
		wantNull bool
	}{
		{"bullet 1+0", "60+0", 1, false},
		{"bullet 2+1", "120+1", 1, false},
		{"blitz 3+0", "180+0", 2, false},
		{"blitz 5+0", "300+0", 2, false},
		{"blitz 3+2", "180+2", 2, false},
		{"rapid 10+0", "600+0", 3, false},
		{"rapid 15+10", "900+10", 3, false},
		{"classical 60+0", "3600+0", 4, false},
		{"classical 90+30", "5400+30", 4, false},
		{"multi-stage 40/5400:1800", "40/5400:1800", 4, false},
		{"missing", "", 0, true},
		{"dash", "-", 0, true},
		{"unparseable", "correspondence", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat, isNull := parseTimeControlCategory(tt.input)
			if isNull != tt.wantNull {
				t.Errorf("isNull = %v, want %v", isNull, tt.wantNull)
			}
			if !isNull && cat != tt.wantCat {
				t.Errorf("category = %d, want %d", cat, tt.wantCat)
			}
		})
	}
}
