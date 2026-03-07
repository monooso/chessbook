package move

import "github.com/monooso/chessbook/chess/board"

// Move represents a chess move.
type Move struct {
	From      board.Square
	To        board.Square
	Promotion board.PieceType // Zero for non-promotion moves.
	Castle    bool            // True for castling moves. To is the rook's square.
}

// promotionChar maps promotion piece types to UCI characters.
var promotionChar = [7]byte{0, 0, 'n', 'b', 'r', 'q', 0}

// String returns the move in UCI notation (e.g. "e2e4", "e7e8q").
func (m Move) String() string {
	s := m.From.String() + m.To.String()
	if m.Promotion != 0 {
		s += string(promotionChar[m.Promotion])
	}
	return s
}
