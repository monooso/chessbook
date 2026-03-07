package move

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
)

// perft counts the number of leaf nodes at a given depth from a position.
// At depth 0, it returns 1 (the position itself).
// At depth 1, it returns the number of legal moves.
// At depth N, it returns the sum of perft(N-1) for each legal move.
func perft(pos *board.Position, depth int) uint64 {
	if depth == 0 {
		return 1
	}

	moves := Generate(pos)
	if depth == 1 {
		return uint64(len(moves))
	}

	var count uint64
	for _, m := range moves {
		after := Apply(pos, m)
		count += perft(&after, depth-1)
	}
	return count
}

// Standard perft positions from the chess programming wiki:
// https://www.chessprogramming.org/Perft_Results

func TestPerftStartingPosition(t *testing.T) {
	p := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")

	tests := []struct {
		depth int
		want  uint64
	}{
		{0, 1},
		{1, 20},
		{2, 400},
		{3, 8902},
		{4, 197281},
	}

	for _, tt := range tests {
		got := perft(&p, tt.depth)
		if got != tt.want {
			t.Errorf("perft(%d) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}

// Position 2: "Kiwipete" — a complex position with many tactical features.
func TestPerftKiwipete(t *testing.T) {
	p := pos(t, "r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq -")

	tests := []struct {
		depth int
		want  uint64
	}{
		{1, 48},
		{2, 2039},
		{3, 97862},
		{4, 4085603},
	}

	for _, tt := range tests {
		got := perft(&p, tt.depth)
		if got != tt.want {
			t.Errorf("perft(%d) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}

// Position 3: tests en passant and promotion edge cases.
func TestPerftPosition3(t *testing.T) {
	p := pos(t, "8/2p5/3p4/KP5r/1R3p1k/8/4P1P1/8 w - -")

	tests := []struct {
		depth int
		want  uint64
	}{
		{1, 14},
		{2, 191},
		{3, 2812},
		{4, 43238},
		{5, 674624},
	}

	for _, tt := range tests {
		got := perft(&p, tt.depth)
		if got != tt.want {
			t.Errorf("perft(%d) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}

// Position 4: mirrored position to test symmetry.
func TestPerftPosition4(t *testing.T) {
	p := pos(t, "r3k2r/Pppp1ppp/1b3nbN/nP6/BBP1P3/q4N2/Pp1P2PP/R2Q1RK1 w kq -")

	tests := []struct {
		depth int
		want  uint64
	}{
		{1, 6},
		{2, 264},
		{3, 9467},
		{4, 422333},
	}

	for _, tt := range tests {
		got := perft(&p, tt.depth)
		if got != tt.want {
			t.Errorf("perft(%d) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}

// Position 5: another complex position.
func TestPerftPosition5(t *testing.T) {
	p := pos(t, "rnbq1k1r/pp1Pbppp/2p5/8/2B5/8/PPP1NnPP/RNBQK2R w KQ -")

	tests := []struct {
		depth int
		want  uint64
	}{
		{1, 44},
		{2, 1486},
		{3, 62379},
		{4, 2103487},
	}

	for _, tt := range tests {
		got := perft(&p, tt.depth)
		if got != tt.want {
			t.Errorf("perft(%d) = %d, want %d", tt.depth, got, tt.want)
		}
	}
}
