package move

import "github.com/monooso/chessbook/chess/board"

// Generate returns all legal moves for the side to move.
func Generate(pos *board.Position) []Move {
	var moves []Move
	color := pos.SideToMove

	for sq := range board.Square(64) {
		p := pos.PieceAt(sq)
		if p.IsEmpty() || p.Color() != color {
			continue
		}

		switch p.Type() {
		case board.Pawn:
			moves = generatePawnMoves(pos, sq, color, moves)
		case board.Knight:
			moves = generateKnightMoves(pos, sq, color, moves)
		case board.Bishop:
			moves = generateSlidingMoves(pos, sq, color, bishopDirs, moves)
		case board.Rook:
			moves = generateSlidingMoves(pos, sq, color, rookDirs, moves)
		case board.Queen:
			moves = generateSlidingMoves(pos, sq, color, queenDirs, moves)
		case board.King:
			moves = generateKingMoves(pos, sq, color, moves)
		}
	}

	moves = generateCastlingMoves(pos, color, moves)

	// Filter to legal moves: remove any move that leaves the king in check.
	legal := moves[:0]
	for _, m := range moves {
		after := Apply(pos, m)
		// Check if our king is in check in the resulting position.
		// After Apply, SideToMove has flipped, so our colour is after.SideToMove.Flip().
		kingSq := board.FindKing(&after, color)
		if kingSq == board.NoSquare {
			continue
		}
		if !IsSquareAttacked(&after, kingSq, color.Flip()) {
			legal = append(legal, m)
		}
	}

	return legal
}

func generatePawnMoves(pos *board.Position, sq board.Square, color board.Color, moves []Move) []Move {
	var forward int8
	var startRank, promoRank board.Rank
	if color == board.White {
		forward = 1
		startRank = board.Rank2
		promoRank = board.Rank8
	} else {
		forward = -1
		startRank = board.Rank7
		promoRank = board.Rank1
	}

	file := int8(sq.File())
	rank := int8(sq.Rank())
	nextRank := rank + forward

	// Single push.
	if nextRank >= 0 && nextRank <= 7 {
		to := board.NewSquare(board.File(file), board.Rank(nextRank))
		if pos.PieceAt(to).IsEmpty() {
			if board.Rank(nextRank) == promoRank {
				moves = addPromotions(sq, to, moves)
			} else {
				moves = append(moves, Move{From: sq, To: to})
			}

			// Double push from starting rank.
			if sq.Rank() == startRank {
				doubleRank := nextRank + forward
				to2 := board.NewSquare(board.File(file), board.Rank(doubleRank))
				if pos.PieceAt(to2).IsEmpty() {
					moves = append(moves, Move{From: sq, To: to2})
				}
			}
		}
	}

	// Captures (including en passant).
	for _, fd := range []int8{-1, 1} {
		captFile := file + fd
		if captFile < 0 || captFile > 7 || nextRank < 0 || nextRank > 7 {
			continue
		}

		to := board.NewSquare(board.File(captFile), board.Rank(nextRank))
		target := pos.PieceAt(to)

		isCapture := !target.IsEmpty() && target.Color() != color
		isEnPassant := to == pos.EnPassant

		if isCapture || isEnPassant {
			if board.Rank(nextRank) == promoRank {
				moves = addPromotions(sq, to, moves)
			} else {
				moves = append(moves, Move{From: sq, To: to})
			}
		}
	}

	return moves
}

func addPromotions(from, to board.Square, moves []Move) []Move {
	for _, pt := range []board.PieceType{board.Queen, board.Rook, board.Bishop, board.Knight} {
		moves = append(moves, Move{From: from, To: to, Promotion: pt})
	}
	return moves
}

func generateKnightMoves(pos *board.Position, sq board.Square, color board.Color, moves []Move) []Move {
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, off := range knightOffsets {
		f, r := file+off.fileDelta, rank+off.rankDelta
		if f < 0 || f > 7 || r < 0 || r > 7 {
			continue
		}

		to := board.NewSquare(board.File(f), board.Rank(r))
		target := pos.PieceAt(to)
		if target.IsEmpty() || target.Color() != color {
			moves = append(moves, Move{From: sq, To: to})
		}
	}

	return moves
}

func generateSlidingMoves(pos *board.Position, sq board.Square, color board.Color, dirs []direction, moves []Move) []Move {
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, dir := range dirs {
		f, r := file+dir.fileDelta, rank+dir.rankDelta
		for f >= 0 && f <= 7 && r >= 0 && r <= 7 {
			to := board.NewSquare(board.File(f), board.Rank(r))
			target := pos.PieceAt(to)

			if target.IsEmpty() {
				moves = append(moves, Move{From: sq, To: to})
			} else {
				if target.Color() != color {
					moves = append(moves, Move{From: sq, To: to})
				}
				break
			}

			f += dir.fileDelta
			r += dir.rankDelta
		}
	}

	return moves
}

func generateKingMoves(pos *board.Position, sq board.Square, color board.Color, moves []Move) []Move {
	file, rank := int8(sq.File()), int8(sq.Rank())

	for _, off := range kingOffsets {
		f, r := file+off.fileDelta, rank+off.rankDelta
		if f < 0 || f > 7 || r < 0 || r > 7 {
			continue
		}

		to := board.NewSquare(board.File(f), board.Rank(r))
		target := pos.PieceAt(to)
		if target.IsEmpty() || target.Color() != color {
			moves = append(moves, Move{From: sq, To: to})
		}
	}

	return moves
}

func generateCastlingMoves(pos *board.Position, color board.Color, moves []Move) []Move {
	kingSq := board.FindKing(pos, color)
	if kingSq == board.NoSquare {
		return moves
	}

	// Can't castle while in check.
	if IsSquareAttacked(pos, kingSq, color.Flip()) {
		return moves
	}

	baseRank := kingSq.Rank()
	kingFile := int8(kingSq.File())

	// Check each file that has castling rights.
	for f := board.FileA; f <= board.FileH; f++ {
		if !pos.Castling.Has(color, f) {
			continue
		}

		rookSq := board.NewSquare(f, baseRank)
		rookFile := int8(f)

		// Determine destination squares.
		var kingDest, rookDest int8
		if rookFile > kingFile {
			// Kingside.
			kingDest = int8(board.FileG)
			rookDest = int8(board.FileF)
		} else {
			// Queenside.
			kingDest = int8(board.FileC)
			rookDest = int8(board.FileD)
		}

		if !canCastle(pos, kingSq, rookSq, kingFile, rookFile, kingDest, rookDest, baseRank, color) {
			continue
		}

		moves = append(moves, Move{From: kingSq, To: rookSq, Castle: true})
	}

	return moves
}

// canCastle checks whether castling is possible: path is clear and king doesn't
// pass through or end in check.
func canCastle(pos *board.Position, kingSq, rookSq board.Square, kingFile, rookFile, kingDest, rookDest int8, rank board.Rank, color board.Color) bool {
	enemy := color.Flip()

	// All squares between king and king's destination (inclusive) must not be attacked.
	// All squares between king and rook must be empty (except king and rook themselves).

	// Check that all squares the king passes through (including destination) are not attacked.
	step := int8(1)
	if kingDest < kingFile {
		step = -1
	}
	for f := kingFile + step; f != kingDest+step; f += step {
		sq := board.NewSquare(board.File(f), rank)
		if IsSquareAttacked(pos, sq, enemy) {
			return false
		}
	}

	// Check that all squares between king and rook are empty.
	minFile := kingFile
	maxFile := rookFile
	if minFile > maxFile {
		minFile, maxFile = maxFile, minFile
	}
	for f := minFile + 1; f < maxFile; f++ {
		sq := board.NewSquare(board.File(f), rank)
		if !pos.PieceAt(sq).IsEmpty() {
			return false
		}
	}

	// Check that the king destination and rook destination are either empty
	// or occupied by the king or rook (Chess960 overlap).
	for _, destFile := range []int8{kingDest, rookDest} {
		sq := board.NewSquare(board.File(destFile), rank)
		p := pos.PieceAt(sq)
		if !p.IsEmpty() && sq != kingSq && sq != rookSq {
			return false
		}
	}

	return true
}
