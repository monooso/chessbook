package board

// File represents a column on the chess board (a through h).
type File int8

const (
	FileA File = iota
	FileB
	FileC
	FileD
	FileE
	FileF
	FileG
	FileH
)

// Rank represents a row on the chess board (1 through 8).
type Rank int8

const (
	Rank1 Rank = iota
	Rank2
	Rank3
	Rank4
	Rank5
	Rank6
	Rank7
	Rank8
)

// Square represents a position on the chess board as an index from 0 to 63.
// Squares are ordered a1=0, b1=1, ..., h1=7, a2=8, ..., h8=63.
type Square int8

const NoSquare Square = -1

// NewSquare creates a square from a file and rank.
func NewSquare(f File, r Rank) Square {
	return Square(int8(r)*8 + int8(f))
}

// File returns the file of the square.
func (s Square) File() File {
	return File(s % 8)
}

// Rank returns the rank of the square.
func (s Square) Rank() Rank {
	return Rank(s / 8)
}

// String returns the algebraic notation for the square (e.g. "e4"), or "-" for NoSquare.
func (s Square) String() string {
	if s == NoSquare {
		return "-"
	}
	return string(rune('a'+s%8)) + string(rune('1'+s/8))
}
