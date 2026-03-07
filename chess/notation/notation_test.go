package notation

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/fen"
	"github.com/monooso/chessbook/chess/move"
)

func pos(t *testing.T, f string) board.Position {
	t.Helper()
	p, err := fen.Parse(f)
	if err != nil {
		t.Fatalf("bad FEN %q: %v", f, err)
	}
	return p
}

func TestSANToMove(t *testing.T) {
	tests := []struct {
		name    string
		fen     string
		san     string
		wantUCI string
	}{
		{
			"pawn single push",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
			"e4", "e2e4",
		},
		{
			"pawn double push",
			"rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3",
			"d5", "d7d5",
		},
		{
			"knight move",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
			"Nf3", "g1f3",
		},
		{
			"pawn capture",
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6",
			"exd5", "e4d5",
		},
		{
			"kingside castling",
			"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq -",
			"O-O", "e1h1",
		},
		{
			"queenside castling",
			"r3kbnr/pppqpppp/2n5/3p1b2/3P1B2/2N5/PPPQPPPP/R3KBNR w KQkq -",
			"O-O-O", "e1a1",
		},
		{
			"promotion",
			"8/4P3/8/8/8/8/8/4K2k w - -",
			"e8=Q", "e7e8q",
		},
		{
			"knight disambiguation by file",
			"8/8/8/8/8/2N1N3/8/4K2k w - -",
			"Nce4", "c3e4",
		},
		{
			"rook disambiguation by rank",
			"4k3/8/8/8/8/R7/8/R3K3 w Q -",
			"R1a2", "a1a2",
		},
		{
			"move with check",
			"rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq -",
			"Qh4+", "d8h4",
		},
		{
			"move with checkmate marker stripped",
			"rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq -",
			"Qh4#", "d8h4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pos(t, tt.fen)
			m, err := SANToMove(&p, tt.san)
			if err != nil {
				t.Fatalf("SANToMove(%q) error: %v", tt.san, err)
			}
			if got := m.String(); got != tt.wantUCI {
				t.Errorf("SANToMove(%q) = %s, want %s", tt.san, got, tt.wantUCI)
			}
		})
	}
}

func TestMoveToSAN(t *testing.T) {
	tests := []struct {
		name string
		fen  string
		move move.Move
		want string
	}{
		{
			"pawn push",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
			move.Move{From: board.NewSquare(board.FileE, board.Rank2), To: board.NewSquare(board.FileE, board.Rank4)},
			"e4",
		},
		{
			"knight move",
			"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -",
			move.Move{From: board.NewSquare(board.FileG, board.Rank1), To: board.NewSquare(board.FileF, board.Rank3)},
			"Nf3",
		},
		{
			"pawn capture",
			"rnbqkbnr/ppp1pppp/8/3p4/4P3/8/PPPP1PPP/RNBQKBNR w KQkq d6",
			move.Move{From: board.NewSquare(board.FileE, board.Rank4), To: board.NewSquare(board.FileD, board.Rank5)},
			"exd5",
		},
		{
			"kingside castling",
			"r1bqk2r/pppp1ppp/2n2n2/2b1p3/2B1P3/5N2/PPPP1PPP/RNBQK2R w KQkq -",
			move.Move{From: board.NewSquare(board.FileE, board.Rank1), To: board.NewSquare(board.FileH, board.Rank1), Castle: true},
			"O-O",
		},
		{
			"queenside castling",
			"r3kbnr/pppqpppp/2n5/3p1b2/3P1B2/2N5/PPPQPPPP/R3KBNR w KQkq -",
			move.Move{From: board.NewSquare(board.FileE, board.Rank1), To: board.NewSquare(board.FileA, board.Rank1), Castle: true},
			"O-O-O",
		},
		{
			"promotion",
			"8/4P3/8/8/8/8/8/4K2k w - -",
			move.Move{From: board.NewSquare(board.FileE, board.Rank7), To: board.NewSquare(board.FileE, board.Rank8), Promotion: board.Queen},
			"e8=Q",
		},
		{
			"check",
			"4k3/8/8/8/8/8/8/4K2R w K -",
			move.Move{From: board.NewSquare(board.FileH, board.Rank1), To: board.NewSquare(board.FileH, board.Rank8)},
			"Rh8+",
		},
		{
			"checkmate",
			"rnbqkbnr/pppp1ppp/8/4p3/6P1/5P2/PPPPP2P/RNBQKBNR b KQkq -",
			move.Move{From: board.NewSquare(board.FileD, board.Rank8), To: board.NewSquare(board.FileH, board.Rank4)},
			"Qh4#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pos(t, tt.fen)
			got := MoveToSAN(&p, tt.move)
			if got != tt.want {
				t.Errorf("MoveToSAN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSANToMoveErrors(t *testing.T) {
	p := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")

	tests := []struct {
		name string
		san  string
	}{
		{"empty string", ""},
		{"invalid piece", "Ze4"},
		{"no matching move", "e5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SANToMove(&p, tt.san)
			if err == nil {
				t.Errorf("SANToMove(%q) expected error, got nil", tt.san)
			}
		})
	}
}
