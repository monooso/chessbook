package board

// CastlingRights tracks which rook files retain castling rights for each colour.
// It is a pair of bitmasks (1 bit per file) indexed by Color, which supports both
// standard chess (rooks on a- and h-files) and Chess960 (rooks on arbitrary files).
// The representation maps directly to the 16 Zobrist castling keys in design doc 011.
type CastlingRights [2]uint8

// Has returns true if the given colour has castling rights for a rook on the given file.
func (cr CastlingRights) Has(c Color, f File) bool {
	return cr[c]&(1<<f) != 0
}

// Set grants castling rights for the given colour and file.
func (cr *CastlingRights) Set(c Color, f File) {
	cr[c] |= 1 << f
}

// Clear removes castling rights for the given colour and file.
func (cr *CastlingRights) Clear(c Color, f File) {
	cr[c] &^= 1 << f
}

// ClearColor removes all castling rights for the given colour.
func (cr *CastlingRights) ClearColor(c Color) {
	cr[c] = 0
}

// Position represents a full chess position: piece placement, side to move,
// castling rights, and en passant square. It does not include the halfmove
// clock or fullmove number, which are excluded from position identity
// (see design doc 003).
type Position struct {
	Squares    [64]Piece
	SideToMove Color
	Castling   CastlingRights
	EnPassant  Square
}

// EmptyPosition returns a position with no pieces, white to move, no castling
// rights, and no en passant square.
func EmptyPosition() Position {
	return Position{EnPassant: NoSquare}
}

// PieceAt returns the piece on the given square.
func (pos *Position) PieceAt(sq Square) Piece {
	return pos.Squares[sq]
}

// Put places a piece on the given square.
func (pos *Position) Put(p Piece, sq Square) {
	pos.Squares[sq] = p
}

// Remove removes the piece from the given square.
func (pos *Position) Remove(sq Square) {
	pos.Squares[sq] = NoPiece
}

// FindKing returns the square of the king for the given colour,
// or NoSquare if no king is found.
func FindKing(pos *Position, c Color) Square {
	king := NewPiece(c, King)
	for sq := Square(0); sq < 64; sq++ {
		if pos.PieceAt(sq) == king {
			return sq
		}
	}
	return NoSquare
}

// StartingPosition returns the standard chess starting position.
func StartingPosition() Position {
	pos := EmptyPosition()

	// White pieces.
	pos.Put(NewPiece(White, Rook), NewSquare(FileA, Rank1))
	pos.Put(NewPiece(White, Knight), NewSquare(FileB, Rank1))
	pos.Put(NewPiece(White, Bishop), NewSquare(FileC, Rank1))
	pos.Put(NewPiece(White, Queen), NewSquare(FileD, Rank1))
	pos.Put(NewPiece(White, King), NewSquare(FileE, Rank1))
	pos.Put(NewPiece(White, Bishop), NewSquare(FileF, Rank1))
	pos.Put(NewPiece(White, Knight), NewSquare(FileG, Rank1))
	pos.Put(NewPiece(White, Rook), NewSquare(FileH, Rank1))
	for f := FileA; f <= FileH; f++ {
		pos.Put(NewPiece(White, Pawn), NewSquare(f, Rank2))
	}

	// Black pieces.
	pos.Put(NewPiece(Black, Rook), NewSquare(FileA, Rank8))
	pos.Put(NewPiece(Black, Knight), NewSquare(FileB, Rank8))
	pos.Put(NewPiece(Black, Bishop), NewSquare(FileC, Rank8))
	pos.Put(NewPiece(Black, Queen), NewSquare(FileD, Rank8))
	pos.Put(NewPiece(Black, King), NewSquare(FileE, Rank8))
	pos.Put(NewPiece(Black, Bishop), NewSquare(FileF, Rank8))
	pos.Put(NewPiece(Black, Knight), NewSquare(FileG, Rank8))
	pos.Put(NewPiece(Black, Rook), NewSquare(FileH, Rank8))
	for f := FileA; f <= FileH; f++ {
		pos.Put(NewPiece(Black, Pawn), NewSquare(f, Rank7))
	}

	// Standard castling rights.
	pos.Castling.Set(White, FileA)
	pos.Castling.Set(White, FileH)
	pos.Castling.Set(Black, FileA)
	pos.Castling.Set(Black, FileH)

	return pos
}
