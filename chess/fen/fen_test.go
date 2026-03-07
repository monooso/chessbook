package fen

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
)

func TestParseStartingPosition(t *testing.T) {
	pos, err := Parse("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := board.StartingPosition()

	// Check all 64 squares.
	for sq := range board.Square(64) {
		if got, w := pos.PieceAt(sq), want.PieceAt(sq); got != w {
			t.Errorf("PieceAt(%s) = %s, want %s", sq, got, w)
		}
	}

	if pos.SideToMove != board.White {
		t.Errorf("SideToMove = %d, want White", pos.SideToMove)
	}

	if pos.Castling != want.Castling {
		t.Errorf("Castling = %v, want %v", pos.Castling, want.Castling)
	}

	if pos.EnPassant != board.NoSquare {
		t.Errorf("EnPassant = %s, want NoSquare", pos.EnPassant)
	}
}

func TestParseAfter1e4(t *testing.T) {
	pos, err := Parse("rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3 0 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pos.SideToMove != board.Black {
		t.Errorf("SideToMove = %d, want Black", pos.SideToMove)
	}

	e4 := board.NewSquare(board.FileE, board.Rank4)
	if got := pos.PieceAt(e4); got != board.NewPiece(board.White, board.Pawn) {
		t.Errorf("PieceAt(e4) = %s, want white pawn", got)
	}

	e2 := board.NewSquare(board.FileE, board.Rank2)
	if got := pos.PieceAt(e2); !got.IsEmpty() {
		t.Errorf("PieceAt(e2) = %s, want empty", got)
	}

	e3 := board.NewSquare(board.FileE, board.Rank3)
	if pos.EnPassant != e3 {
		t.Errorf("EnPassant = %s, want %s", pos.EnPassant, e3)
	}
}

func TestParseNoCastlingRights(t *testing.T) {
	pos, err := Parse("8/8/8/8/8/8/8/4K2k w - - 0 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var empty board.CastlingRights
	if pos.Castling != empty {
		t.Errorf("Castling = %v, want no rights", pos.Castling)
	}
}

func TestParsePartialCastlingRights(t *testing.T) {
	pos, err := Parse("r3k2r/8/8/8/8/8/8/R3K2R w Kq - 0 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pos.Castling.Has(board.White, board.FileH) {
		t.Error("expected white kingside castling right")
	}
	if pos.Castling.Has(board.White, board.FileA) {
		t.Error("unexpected white queenside castling right")
	}
	if !pos.Castling.Has(board.Black, board.FileA) {
		t.Error("expected black queenside castling right")
	}
	if pos.Castling.Has(board.Black, board.FileH) {
		t.Error("unexpected black kingside castling right")
	}
}

func TestParseChess960Castling(t *testing.T) {
	// Chess960 position with rooks on b and g files.
	pos, err := Parse("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w BGbg - 0 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pos.Castling.Has(board.White, board.FileB) {
		t.Error("expected white b-file castling right")
	}
	if !pos.Castling.Has(board.White, board.FileG) {
		t.Error("expected white g-file castling right")
	}
	if !pos.Castling.Has(board.Black, board.FileB) {
		t.Error("expected black b-file castling right")
	}
	if !pos.Castling.Has(board.Black, board.FileG) {
		t.Error("expected black g-file castling right")
	}
}

func TestParseOptionalFields(t *testing.T) {
	// FEN with only 4 fields (no halfmove clock or fullmove number).
	pos, err := Parse("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pos.SideToMove != board.White {
		t.Errorf("SideToMove = %d, want White", pos.SideToMove)
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"empty string", ""},
		{"too few fields", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq"},
		{"bad active color", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR x KQkq -"},
		{"too few ranks", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP w KQkq -"},
		{"too many squares in rank", "rnbqkbnrr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -"},
		{"invalid piece character", "xnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -"},
		{"bad en passant square", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq z9"},
		{"invalid castling character", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQxq -"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.fen)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		fen  string
	}{
		{"starting position", "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -"},
		{"after 1.e4", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3"},
		{"no castling", "8/8/8/8/8/8/8/4K2k w - -"},
		{"partial castling", "r3k2r/8/8/8/8/8/8/R3K2R w Kq -"},
		{"black to move", "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pos, err := Parse(tt.fen)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tt.fen, err)
			}

			got := Format(pos)
			if got != tt.fen {
				t.Errorf("Format(Parse(%q)) = %q", tt.fen, got)
			}
		})
	}
}

func TestFormatChess960Castling(t *testing.T) {
	pos, err := Parse("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w BGbg -")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := Format(pos)
	if got != "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w BGbg -" {
		t.Errorf("Format = %q, want Chess960 castling notation", got)
	}
}
