package board

import "testing"

func TestEncodeDecodeRoundTrip(t *testing.T) {
	positions := []struct {
		name string
		pos  Position
	}{
		{"empty", EmptyPosition()},
		{"starting", StartingPosition()},
	}

	for _, tt := range positions {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Encode(&tt.pos)
			decoded := Decode(encoded)

			if decoded != tt.pos {
				t.Errorf("round-trip failed for %s", tt.name)
			}
		})
	}
}

func TestEncodeSize(t *testing.T) {
	pos := StartingPosition()
	encoded := Encode(&pos)
	if len(encoded) != EncodedSize {
		t.Errorf("encoded size = %d, want %d", len(encoded), EncodedSize)
	}
}

func TestEncodeDeterministic(t *testing.T) {
	pos := StartingPosition()
	e1 := Encode(&pos)
	e2 := Encode(&pos)
	if e1 != e2 {
		t.Error("same position produced different encodings")
	}
}

func TestEncodeDifferentPositions(t *testing.T) {
	p1 := StartingPosition()
	p2 := StartingPosition()
	p2.SideToMove = Black

	e1 := Encode(&p1)
	e2 := Encode(&p2)
	if e1 == e2 {
		t.Error("different positions produced the same encoding")
	}
}

func TestEncodeCastlingRights(t *testing.T) {
	p1 := EmptyPosition()
	p1.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	p1.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	p1.Castling.Set(White, FileH)

	p2 := EmptyPosition()
	p2.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	p2.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))

	e1 := Encode(&p1)
	e2 := Encode(&p2)
	if e1 == e2 {
		t.Error("different castling rights produced the same encoding")
	}
}

func TestEncodeCastlingRightsChess960(t *testing.T) {
	// Chess960: castling rights on non-standard files.
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileD, Rank1))
	pos.Put(NewPiece(White, Rook), NewSquare(FileB, Rank1))
	pos.Put(NewPiece(White, Rook), NewSquare(FileF, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Castling.Set(White, FileB)
	pos.Castling.Set(White, FileF)

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if !decoded.Castling.Has(White, FileB) {
		t.Error("lost white b-file castling right")
	}
	if !decoded.Castling.Has(White, FileF) {
		t.Error("lost white f-file castling right")
	}
}

func TestEncodeEnPassantBlackToMove(t *testing.T) {
	// After 1.d4, black to move, en passant on d3. Black pawn on e4 can capture.
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Put(NewPiece(Black, Pawn), NewSquare(FileE, Rank4))
	pos.SideToMove = Black
	pos.EnPassant = NewSquare(FileD, Rank3)

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if decoded.EnPassant != NewSquare(FileD, Rank3) {
		t.Errorf("EnPassant = %s, want d3", decoded.EnPassant)
	}
}

func TestEncodeEnPassantWhiteToMove(t *testing.T) {
	// After 1.e4 d5 2.e5 f5, white to move, en passant on f6. White pawn on e5 can capture.
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Put(NewPiece(White, Pawn), NewSquare(FileE, Rank5))
	pos.SideToMove = White
	pos.EnPassant = NewSquare(FileF, Rank6)

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if decoded.EnPassant != NewSquare(FileF, Rank6) {
		t.Errorf("EnPassant = %s, want f6", decoded.EnPassant)
	}
}

func TestEncodeNoEnPassant(t *testing.T) {
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if decoded.EnPassant != NoSquare {
		t.Errorf("EnPassant = %s, want NoSquare", decoded.EnPassant)
	}
}

func TestEncodeEnPassantNormalized(t *testing.T) {
	// EP square set, but no opposing pawn can capture. The encoding should
	// strip the EP square to match Zobrist normalisation (design doc 003).
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Put(NewPiece(White, Pawn), NewSquare(FileD, Rank5))
	pos.SideToMove = White
	pos.EnPassant = NewSquare(FileF, Rank6) // no white pawn on e5 or g5

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if decoded.EnPassant != NoSquare {
		t.Errorf("phantom EP not stripped: got %s, want NoSquare", decoded.EnPassant)
	}
}

func TestEncodeEnPassantPreservedWhenLegal(t *testing.T) {
	// EP square set, and an opposing pawn CAN capture. The encoding should
	// preserve the EP square.
	pos := EmptyPosition()
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Put(NewPiece(White, Pawn), NewSquare(FileE, Rank5)) // can capture on f6
	pos.SideToMove = White
	pos.EnPassant = NewSquare(FileF, Rank6)

	encoded := Encode(&pos)
	decoded := Decode(encoded)

	if decoded.EnPassant != NewSquare(FileF, Rank6) {
		t.Errorf("legal EP stripped: got %s, want f6", decoded.EnPassant)
	}
}
