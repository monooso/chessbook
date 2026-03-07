package move

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/fen"
)

func pos(t *testing.T, f string) board.Position {
	t.Helper()
	p, err := fen.Parse(f)
	if err != nil {
		t.Fatalf("bad FEN %q: %v", f, err)
	}
	return p
}

func TestIsSquareAttacked(t *testing.T) {
	tests := []struct {
		name     string
		fen      string
		sq       board.Square
		byColor  board.Color
		attacked bool
	}{
		{
			"pawn attacks diagonally",
			"8/8/8/8/8/3P4/8/8 w - -",
			board.NewSquare(board.FileE, board.Rank4),
			board.White,
			true,
		},
		{
			"pawn does not attack forward",
			"8/8/8/8/8/3P4/8/8 w - -",
			board.NewSquare(board.FileD, board.Rank4),
			board.White,
			false,
		},
		{
			"black pawn attacks diagonally downward",
			"8/8/3p4/8/8/8/8/8 w - -",
			board.NewSquare(board.FileC, board.Rank5),
			board.Black,
			true,
		},
		{
			"knight attacks",
			"8/8/8/8/3N4/8/8/8 w - -",
			board.NewSquare(board.FileE, board.Rank6),
			board.White,
			true,
		},
		{
			"knight does not attack adjacent",
			"8/8/8/8/3N4/8/8/8 w - -",
			board.NewSquare(board.FileD, board.Rank5),
			board.White,
			false,
		},
		{
			"bishop attacks diagonally",
			"8/8/8/8/3B4/8/8/8 w - -",
			board.NewSquare(board.FileF, board.Rank6),
			board.White,
			true,
		},
		{
			"bishop blocked by piece",
			"8/8/8/4p3/3B4/8/8/8 w - -",
			board.NewSquare(board.FileF, board.Rank6),
			board.White,
			false,
		},
		{
			"rook attacks along file",
			"8/8/8/8/3R4/8/8/8 w - -",
			board.NewSquare(board.FileD, board.Rank8),
			board.White,
			true,
		},
		{
			"rook attacks along rank",
			"8/8/8/8/3R4/8/8/8 w - -",
			board.NewSquare(board.FileH, board.Rank4),
			board.White,
			true,
		},
		{
			"rook blocked by piece",
			"8/8/8/3p4/3R4/8/8/8 w - -",
			board.NewSquare(board.FileD, board.Rank8),
			board.White,
			false,
		},
		{
			"queen attacks diagonally",
			"8/8/8/8/3Q4/8/8/8 w - -",
			board.NewSquare(board.FileF, board.Rank6),
			board.White,
			true,
		},
		{
			"queen attacks along rank",
			"8/8/8/8/3Q4/8/8/8 w - -",
			board.NewSquare(board.FileA, board.Rank4),
			board.White,
			true,
		},
		{
			"king attacks adjacent",
			"8/8/8/8/3K4/8/8/8 w - -",
			board.NewSquare(board.FileE, board.Rank5),
			board.White,
			true,
		},
		{
			"king does not attack two squares away",
			"8/8/8/8/3K4/8/8/8 w - -",
			board.NewSquare(board.FileF, board.Rank6),
			board.White,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pos(t, tt.fen)
			got := IsSquareAttacked(&p, tt.sq, tt.byColor)
			if got != tt.attacked {
				t.Errorf("IsSquareAttacked(%s, %s) = %v, want %v", tt.sq, tt.fen, got, tt.attacked)
			}
		})
	}
}

func TestIsInCheck(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		inCheck bool
	}{
		{
			"starting position not in check",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
			false,
		},
		{
			"white king in check from bishop",
			"8/8/8/8/8/8/4b3/3K4 w - -",
			true,
		},
		{
			"black king in check from rook",
			"3R4/8/8/8/8/8/8/3k4 b - -",
			true,
		},
		{
			"not in check - wrong color attacking",
			"8/8/8/8/5B2/8/8/3K4 w - -",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pos(t, tt.fen)
			got := IsInCheck(&p)
			if got != tt.inCheck {
				t.Errorf("IsInCheck() = %v, want %v", got, tt.inCheck)
			}
		})
	}
}
