package board

import "testing"

func TestEmptyPosition(t *testing.T) {
	pos := EmptyPosition()

	// All squares should be empty.
	for sq := Square(0); sq < 64; sq++ {
		if p := pos.PieceAt(sq); !p.IsEmpty() {
			t.Errorf("PieceAt(%s) = %s, want empty", sq, p)
		}
	}

	if pos.SideToMove != White {
		t.Errorf("SideToMove = %d, want White", pos.SideToMove)
	}

	if pos.Castling != [2]uint8{} {
		t.Errorf("Castling = %v, want zero value", pos.Castling)
	}

	if pos.EnPassant != NoSquare {
		t.Errorf("EnPassant = %s, want NoSquare", pos.EnPassant)
	}
}

func TestPutAndPieceAt(t *testing.T) {
	pos := EmptyPosition()
	wk := NewPiece(White, King)
	e1 := NewSquare(FileE, Rank1)

	pos.Put(wk, e1)

	if got := pos.PieceAt(e1); got != wk {
		t.Errorf("PieceAt(e1) = %s, want %s", got, wk)
	}
}

func TestRemove(t *testing.T) {
	pos := EmptyPosition()
	wk := NewPiece(White, King)
	e1 := NewSquare(FileE, Rank1)

	pos.Put(wk, e1)
	pos.Remove(e1)

	if got := pos.PieceAt(e1); !got.IsEmpty() {
		t.Errorf("PieceAt(e1) after Remove = %s, want empty", got)
	}
}

func TestStartingPosition(t *testing.T) {
	pos := StartingPosition()

	// Check a few key pieces.
	tests := []struct {
		sq   Square
		want Piece
	}{
		{NewSquare(FileA, Rank1), NewPiece(White, Rook)},
		{NewSquare(FileE, Rank1), NewPiece(White, King)},
		{NewSquare(FileD, Rank1), NewPiece(White, Queen)},
		{NewSquare(FileA, Rank2), NewPiece(White, Pawn)},
		{NewSquare(FileH, Rank2), NewPiece(White, Pawn)},
		{NewSquare(FileA, Rank7), NewPiece(Black, Pawn)},
		{NewSquare(FileE, Rank8), NewPiece(Black, King)},
		{NewSquare(FileD, Rank8), NewPiece(Black, Queen)},
		{NewSquare(FileH, Rank8), NewPiece(Black, Rook)},
		// Empty squares in the middle.
		{NewSquare(FileE, Rank4), NoPiece},
		{NewSquare(FileD, Rank5), NoPiece},
	}

	for _, tt := range tests {
		t.Run(tt.sq.String(), func(t *testing.T) {
			if got := pos.PieceAt(tt.sq); got != tt.want {
				t.Errorf("PieceAt(%s) = %s, want %s", tt.sq, got, tt.want)
			}
		})
	}

	if pos.SideToMove != White {
		t.Errorf("SideToMove = %d, want White", pos.SideToMove)
	}

	// Standard starting position: all four castling rights.
	if !pos.Castling.Has(White, FileA) {
		t.Error("expected white queenside castling right")
	}
	if !pos.Castling.Has(White, FileH) {
		t.Error("expected white kingside castling right")
	}
	if !pos.Castling.Has(Black, FileA) {
		t.Error("expected black queenside castling right")
	}
	if !pos.Castling.Has(Black, FileH) {
		t.Error("expected black kingside castling right")
	}

	if pos.EnPassant != NoSquare {
		t.Errorf("EnPassant = %s, want NoSquare", pos.EnPassant)
	}
}

func TestCastlingRights(t *testing.T) {
	var cr CastlingRights

	// Initially empty.
	if cr.Has(White, FileH) {
		t.Error("empty CastlingRights should not have any rights")
	}

	// Set and check.
	cr.Set(White, FileH)
	if !cr.Has(White, FileH) {
		t.Error("Has(White, FileH) = false after Set")
	}
	if cr.Has(Black, FileH) {
		t.Error("Has(Black, FileH) = true, should only affect White")
	}

	// Clear.
	cr.Clear(White, FileH)
	if cr.Has(White, FileH) {
		t.Error("Has(White, FileH) = true after Clear")
	}

	// Chess960: castling rights on non-standard files.
	cr.Set(White, FileC)
	cr.Set(White, FileF)
	if !cr.Has(White, FileC) {
		t.Error("Has(White, FileC) = false after Set")
	}
	if !cr.Has(White, FileF) {
		t.Error("Has(White, FileF) = false after Set")
	}
}

func TestCastlingRightsClearColor(t *testing.T) {
	var cr CastlingRights
	cr.Set(White, FileA)
	cr.Set(White, FileH)
	cr.Set(Black, FileA)
	cr.Set(Black, FileH)

	cr.ClearColor(White)

	if cr.Has(White, FileA) || cr.Has(White, FileH) {
		t.Error("white castling rights should be cleared")
	}
	if !cr.Has(Black, FileA) || !cr.Has(Black, FileH) {
		t.Error("black castling rights should be unaffected")
	}
}
