package board

import "testing"

func TestSquareFromFileRank(t *testing.T) {
	tests := []struct {
		name string
		file File
		rank Rank
		want Square
	}{
		{"a1", FileA, Rank1, 0},
		{"b1", FileB, Rank1, 1},
		{"h1", FileH, Rank1, 7},
		{"a2", FileA, Rank2, 8},
		{"e4", FileE, Rank4, 28},
		{"h8", FileH, Rank8, 63},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSquare(tt.file, tt.rank)
			if got != tt.want {
				t.Errorf("NewSquare(%d, %d) = %d, want %d", tt.file, tt.rank, got, tt.want)
			}
		})
	}
}

func TestSquareFile(t *testing.T) {
	tests := []struct {
		sq   Square
		want File
	}{
		{0, FileA},
		{1, FileB},
		{7, FileH},
		{8, FileA},
		{28, FileE},
		{63, FileH},
	}

	for _, tt := range tests {
		t.Run(tt.sq.String(), func(t *testing.T) {
			if got := tt.sq.File(); got != tt.want {
				t.Errorf("Square(%d).File() = %d, want %d", tt.sq, got, tt.want)
			}
		})
	}
}

func TestSquareRank(t *testing.T) {
	tests := []struct {
		sq   Square
		want Rank
	}{
		{0, Rank1},
		{1, Rank1},
		{7, Rank1},
		{8, Rank2},
		{28, Rank4},
		{63, Rank8},
	}

	for _, tt := range tests {
		t.Run(tt.sq.String(), func(t *testing.T) {
			if got := tt.sq.Rank(); got != tt.want {
				t.Errorf("Square(%d).Rank() = %d, want %d", tt.sq, got, tt.want)
			}
		})
	}
}

func TestSquareString(t *testing.T) {
	tests := []struct {
		sq   Square
		want string
	}{
		{0, "a1"},
		{1, "b1"},
		{7, "h1"},
		{8, "a2"},
		{28, "e4"},
		{63, "h8"},
		{NoSquare, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.sq.String(); got != tt.want {
				t.Errorf("Square(%d).String() = %q, want %q", tt.sq, got, tt.want)
			}
		})
	}
}
