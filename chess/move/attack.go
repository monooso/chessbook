package move

import "github.com/monooso/chessbook/chess/board"

// IsSquareAttacked returns true if the given square is attacked by any piece
// of the given colour.
func IsSquareAttacked(pos *board.Position, sq board.Square, byColor board.Color) bool {
	return isPawnAttacking(pos, sq, byColor) ||
		isKnightAttacking(pos, sq, byColor) ||
		isSlidingAttacking(pos, sq, byColor, board.Bishop, bishopDirs) ||
		isSlidingAttacking(pos, sq, byColor, board.Rook, rookDirs) ||
		isSlidingAttacking(pos, sq, byColor, board.Queen, queenDirs) ||
		isKingAttacking(pos, sq, byColor)
}

// IsInCheck returns true if the side to move is in check.
func IsInCheck(pos *board.Position) bool {
	kingSq := findKing(pos, pos.SideToMove)
	if kingSq == board.NoSquare {
		return false
	}
	return IsSquareAttacked(pos, kingSq, pos.SideToMove.Flip())
}

// Direction offsets for sliding pieces.
type direction struct {
	fileDelta int8
	rankDelta int8
}

var bishopDirs = []direction{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}}
var rookDirs = []direction{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
var queenDirs = append(append([]direction{}, bishopDirs...), rookDirs...)

// Knight move offsets.
var knightOffsets = []direction{
	{1, 2}, {2, 1}, {2, -1}, {1, -2},
	{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2},
}

// King move offsets (same as queen directions, one step).
var kingOffsets = queenDirs

func isPawnAttacking(pos *board.Position, sq board.Square, byColor board.Color) bool {
	// Pawns attack diagonally. A pawn of byColor attacks sq if there is a
	// pawn on a diagonal square one rank "behind" sq from the pawn's perspective.
	var rankBehind int8
	if byColor == board.White {
		rankBehind = -1 // White pawns attack upward, so look one rank below.
	} else {
		rankBehind = 1 // Black pawns attack downward, so look one rank above.
	}

	targetRank := int8(sq.Rank()) + rankBehind
	if targetRank < 0 || targetRank > 7 {
		return false
	}

	pawn := board.NewPiece(byColor, board.Pawn)
	file := int8(sq.File())

	for _, fd := range []int8{-1, 1} {
		f := file + fd
		if f < 0 || f > 7 {
			continue
		}
		from := board.NewSquare(board.File(f), board.Rank(targetRank))
		if pos.PieceAt(from) == pawn {
			return true
		}
	}

	return false
}

func isKnightAttacking(pos *board.Position, sq board.Square, byColor board.Color) bool {
	knight := board.NewPiece(byColor, board.Knight)
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, off := range knightOffsets {
		f, r := file+off.fileDelta, rank+off.rankDelta
		if f < 0 || f > 7 || r < 0 || r > 7 {
			continue
		}
		if pos.PieceAt(board.NewSquare(board.File(f), board.Rank(r))) == knight {
			return true
		}
	}

	return false
}

func isSlidingAttacking(pos *board.Position, sq board.Square, byColor board.Color, pt board.PieceType, dirs []direction) bool {
	piece := board.NewPiece(byColor, pt)
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, dir := range dirs {
		f, r := file+dir.fileDelta, rank+dir.rankDelta
		for f >= 0 && f <= 7 && r >= 0 && r <= 7 {
			p := pos.PieceAt(board.NewSquare(board.File(f), board.Rank(r)))
			if p == piece {
				return true
			}
			if !p.IsEmpty() {
				break // Blocked by another piece.
			}
			f += dir.fileDelta
			r += dir.rankDelta
		}
	}

	return false
}

func isKingAttacking(pos *board.Position, sq board.Square, byColor board.Color) bool {
	king := board.NewPiece(byColor, board.King)
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, off := range kingOffsets {
		f, r := file+off.fileDelta, rank+off.rankDelta
		if f < 0 || f > 7 || r < 0 || r > 7 {
			continue
		}
		if pos.PieceAt(board.NewSquare(board.File(f), board.Rank(r))) == king {
			return true
		}
	}

	return false
}

// findKing returns the square of the king of the given colour, or NoSquare if not found.
func findKing(pos *board.Position, c board.Color) board.Square {
	king := board.NewPiece(c, board.King)
	for sq := board.Square(0); sq < 64; sq++ {
		if pos.PieceAt(sq) == king {
			return sq
		}
	}
	return board.NoSquare
}
