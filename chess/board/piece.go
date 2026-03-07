package board

// Color represents the colour of a piece or side to move.
type Color int8

const (
	White Color = 0
	Black Color = 1
)

// Flip returns the opposite colour.
func (c Color) Flip() Color {
	return 1 - c
}

// PieceType represents the type of a chess piece.
type PieceType int8

const (
	Pawn   PieceType = 1
	Knight PieceType = 2
	Bishop PieceType = 3
	Rook   PieceType = 4
	Queen  PieceType = 5
	King   PieceType = 6
)

// Piece represents a chess piece with colour and type.
// The zero value is NoPiece (an empty square).
// Encoding: White pieces are 1–6, Black pieces are 7–12.
// This matches the nibble encoding in design doc 010.
type Piece int8

const NoPiece Piece = 0

// NewPiece creates a piece from a colour and piece type.
func NewPiece(c Color, pt PieceType) Piece {
	return Piece(int8(c)*6 + int8(pt))
}

// IsEmpty returns true if the piece is NoPiece.
func (p Piece) IsEmpty() bool {
	return p == NoPiece
}

// Color returns the colour of the piece. Undefined for NoPiece.
func (p Piece) Color() Color {
	return Color((p - 1) / 6)
}

// Type returns the piece type. Undefined for NoPiece.
func (p Piece) Type() PieceType {
	return PieceType((p-1)%6 + 1)
}

// String returns a single-character representation of the piece.
// White pieces are uppercase (PNBRQK), black pieces are lowercase (pnbrqk),
// and NoPiece is ".".
func (p Piece) String() string {
	const symbols = ".PNBRQKpnbrqk"
	if p < 0 || int(p) >= len(symbols) {
		return "?"
	}
	return string(symbols[p])
}
