package zobrist

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

func TestHashDeterministic(t *testing.T) {
	p := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	h1 := Hash(&p)
	h2 := Hash(&p)
	if h1 != h2 {
		t.Errorf("same position produced different hashes: %x vs %x", h1, h2)
	}
}

func TestHashDifferentPositions(t *testing.T) {
	p1 := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	p2 := pos(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3")

	h1 := Hash(&p1)
	h2 := Hash(&p2)
	if h1 == h2 {
		t.Error("different positions produced the same hash")
	}
}

func TestHashSideToMoveMatters(t *testing.T) {
	// Same piece placement but different side to move.
	p1 := pos(t, "8/8/8/8/8/8/8/4K2k w - -")
	p2 := pos(t, "8/8/8/8/8/8/8/4K2k b - -")

	h1 := Hash(&p1)
	h2 := Hash(&p2)
	if h1 == h2 {
		t.Error("different side to move produced the same hash")
	}
}

func TestHashCastlingRightsMatter(t *testing.T) {
	p1 := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq -")
	p2 := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w Kq -")

	h1 := Hash(&p1)
	h2 := Hash(&p2)
	if h1 == h2 {
		t.Error("different castling rights produced the same hash")
	}
}

func TestHashEnPassantNormalization(t *testing.T) {
	// Position with en passant square e3 but no black pawn can capture.
	// The en passant should be ignored in the hash.
	p1 := pos(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3")
	p2 := pos(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq -")

	h1 := Hash(&p1)
	h2 := Hash(&p2)
	if h1 != h2 {
		t.Errorf("phantom en passant should not affect hash: %x vs %x", h1, h2)
	}
}

func TestHashEnPassantIncludedWhenCapturePossible(t *testing.T) {
	// Black pawn on d4, white pawn just pushed e2-e4. En passant on e3 is legal.
	p1 := pos(t, "8/8/8/8/3pP3/8/8/4K2k b - e3")
	p2 := pos(t, "8/8/8/8/3pP3/8/8/4K2k b - -")

	h1 := Hash(&p1)
	h2 := Hash(&p2)
	if h1 == h2 {
		t.Error("legal en passant should affect hash")
	}
}

func TestHashNonZero(t *testing.T) {
	// Empty board with just kings should still produce a non-zero hash.
	p := pos(t, "8/8/8/8/8/8/8/4K2k w - -")
	h := Hash(&p)
	if h == 0 {
		t.Error("hash should not be zero")
	}
}

func TestHashTableSize(t *testing.T) {
	// Verify the random table has the expected number of entries.
	// 12 pieces × 64 squares + 16 castling + 8 en passant + 1 side to move = 793
	expected := 12*64 + 16 + 8 + 1
	if len(table) != expected {
		t.Errorf("table has %d entries, want %d", len(table), expected)
	}
}
