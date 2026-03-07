package move

import (
	"testing"

	"github.com/monooso/chessbook/chess/board"
)

func TestApplySimpleMove(t *testing.T) {
	p := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	m := Move{
		From: board.NewSquare(board.FileE, board.Rank2),
		To:   board.NewSquare(board.FileE, board.Rank4),
	}

	result := Apply(&p, m)

	// Pawn should be on e4.
	e4 := board.NewSquare(board.FileE, board.Rank4)
	if got := result.PieceAt(e4); got != board.NewPiece(board.White, board.Pawn) {
		t.Errorf("PieceAt(e4) = %s, want white pawn", got)
	}

	// e2 should be empty.
	e2 := board.NewSquare(board.FileE, board.Rank2)
	if got := result.PieceAt(e2); !got.IsEmpty() {
		t.Errorf("PieceAt(e2) = %s, want empty", got)
	}

	// Side to move should flip.
	if result.SideToMove != board.Black {
		t.Errorf("SideToMove = %d, want Black", result.SideToMove)
	}
}

func TestApplyEnPassantSquareSet(t *testing.T) {
	p := pos(t, "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
	m := Move{
		From: board.NewSquare(board.FileE, board.Rank2),
		To:   board.NewSquare(board.FileE, board.Rank4),
	}

	result := Apply(&p, m)

	e3 := board.NewSquare(board.FileE, board.Rank3)
	if result.EnPassant != e3 {
		t.Errorf("EnPassant = %s, want %s", result.EnPassant, e3)
	}
}

func TestApplyEnPassantCleared(t *testing.T) {
	// After a non-double-pawn push, en passant should be cleared.
	p := pos(t, "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq e3")
	m := Move{
		From: board.NewSquare(board.FileG, board.Rank8),
		To:   board.NewSquare(board.FileF, board.Rank6),
	}

	result := Apply(&p, m)

	if result.EnPassant != board.NoSquare {
		t.Errorf("EnPassant = %s, want NoSquare", result.EnPassant)
	}
}

func TestApplyEnPassantCapture(t *testing.T) {
	// White pawn on e5, black pawn just moved d7-d5. En passant is d6.
	p := pos(t, "8/8/8/3pP3/8/8/8/4K2k w - d6")
	m := Move{
		From: board.NewSquare(board.FileE, board.Rank5),
		To:   board.NewSquare(board.FileD, board.Rank6),
	}

	result := Apply(&p, m)

	// White pawn should be on d6.
	d6 := board.NewSquare(board.FileD, board.Rank6)
	if got := result.PieceAt(d6); got != board.NewPiece(board.White, board.Pawn) {
		t.Errorf("PieceAt(d6) = %s, want white pawn", got)
	}

	// The captured pawn on d5 should be gone.
	d5 := board.NewSquare(board.FileD, board.Rank5)
	if got := result.PieceAt(d5); !got.IsEmpty() {
		t.Errorf("PieceAt(d5) = %s, want empty (captured en passant)", got)
	}
}

func TestApplyPromotion(t *testing.T) {
	p := pos(t, "8/4P3/8/8/8/8/8/4K2k w - -")
	m := Move{
		From:      board.NewSquare(board.FileE, board.Rank7),
		To:        board.NewSquare(board.FileE, board.Rank8),
		Promotion: board.Queen,
	}

	result := Apply(&p, m)

	e8 := board.NewSquare(board.FileE, board.Rank8)
	if got := result.PieceAt(e8); got != board.NewPiece(board.White, board.Queen) {
		t.Errorf("PieceAt(e8) = %s, want white queen", got)
	}
}

func TestApplyCastlingKingside(t *testing.T) {
	p := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq -")
	m := Move{
		From:   board.NewSquare(board.FileE, board.Rank1),
		To:     board.NewSquare(board.FileH, board.Rank1),
		Castle: true,
	}

	result := Apply(&p, m)

	// King should be on g1.
	g1 := board.NewSquare(board.FileG, board.Rank1)
	if got := result.PieceAt(g1); got != board.NewPiece(board.White, board.King) {
		t.Errorf("PieceAt(g1) = %s, want white king", got)
	}

	// Rook should be on f1.
	f1 := board.NewSquare(board.FileF, board.Rank1)
	if got := result.PieceAt(f1); got != board.NewPiece(board.White, board.Rook) {
		t.Errorf("PieceAt(f1) = %s, want white rook", got)
	}

	// Original squares should be empty.
	e1 := board.NewSquare(board.FileE, board.Rank1)
	h1 := board.NewSquare(board.FileH, board.Rank1)
	if got := result.PieceAt(e1); !got.IsEmpty() {
		t.Errorf("PieceAt(e1) = %s, want empty", got)
	}
	if got := result.PieceAt(h1); !got.IsEmpty() {
		t.Errorf("PieceAt(h1) = %s, want empty", got)
	}

	// All white castling rights should be gone.
	if result.Castling.Has(board.White, board.FileA) || result.Castling.Has(board.White, board.FileH) {
		t.Error("white should have no castling rights after castling")
	}
}

func TestApplyCastlingQueenside(t *testing.T) {
	p := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq -")
	m := Move{
		From:   board.NewSquare(board.FileE, board.Rank1),
		To:     board.NewSquare(board.FileA, board.Rank1),
		Castle: true,
	}

	result := Apply(&p, m)

	// King should be on c1.
	c1 := board.NewSquare(board.FileC, board.Rank1)
	if got := result.PieceAt(c1); got != board.NewPiece(board.White, board.King) {
		t.Errorf("PieceAt(c1) = %s, want white king", got)
	}

	// Rook should be on d1.
	d1 := board.NewSquare(board.FileD, board.Rank1)
	if got := result.PieceAt(d1); got != board.NewPiece(board.White, board.Rook) {
		t.Errorf("PieceAt(d1) = %s, want white rook", got)
	}
}

func TestApplyCastlingRightsLostOnKingMove(t *testing.T) {
	p := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq -")
	m := Move{
		From: board.NewSquare(board.FileE, board.Rank1),
		To:   board.NewSquare(board.FileD, board.Rank1),
	}

	result := Apply(&p, m)

	if result.Castling.Has(board.White, board.FileA) || result.Castling.Has(board.White, board.FileH) {
		t.Error("white should lose all castling rights when king moves")
	}
	// Black rights should be unaffected.
	if !result.Castling.Has(board.Black, board.FileA) || !result.Castling.Has(board.Black, board.FileH) {
		t.Error("black castling rights should be unaffected")
	}
}

func TestApplyCastlingRightsLostOnRookMove(t *testing.T) {
	p := pos(t, "r3k2r/8/8/8/8/8/8/R3K2R w KQkq -")
	m := Move{
		From: board.NewSquare(board.FileH, board.Rank1),
		To:   board.NewSquare(board.FileH, board.Rank2),
	}

	result := Apply(&p, m)

	if result.Castling.Has(board.White, board.FileH) {
		t.Error("white should lose kingside castling right when h-file rook moves")
	}
	if !result.Castling.Has(board.White, board.FileA) {
		t.Error("white should keep queenside castling right")
	}
}

func TestApplyCastlingRightsLostOnRookCapture(t *testing.T) {
	// A piece captures the black rook on h8.
	p := pos(t, "r3k2r/8/8/8/8/8/7B/R3K2R w KQkq -")
	m := Move{
		From: board.NewSquare(board.FileH, board.Rank2),
		To:   board.NewSquare(board.FileH, board.Rank8), // Capture on the rook's corner.
	}

	// This test doesn't use a real bishop move but tests the castling rights update
	// when a piece lands on a corner square where a rook with castling rights sits.
	result := Apply(&p, m)

	if result.Castling.Has(board.Black, board.FileH) {
		t.Error("black should lose kingside castling right when rook is captured")
	}
}
