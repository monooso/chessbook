package move

import "github.com/monooso/chessbook/chess/board"

// Apply applies a move to a position and returns the resulting position.
// The move is assumed to be legal; no validation is performed.
func Apply(pos *board.Position, m Move) board.Position {
	result := *pos
	result.EnPassant = board.NoSquare

	piece := result.PieceAt(m.From)
	color := piece.Color()
	pt := piece.Type()

	if m.Castle {
		applyCastle(&result, m, color)
	} else {
		// Handle en passant capture.
		if pt == board.Pawn && m.To == pos.EnPassant {
			applyEnPassantCapture(&result, m)
		}

		// Move the piece.
		result.Remove(m.From)
		if m.Promotion != 0 {
			result.Put(board.NewPiece(color, m.Promotion), m.To)
		} else {
			result.Put(piece, m.To)
		}

		// Set en passant square for double pawn push.
		if pt == board.Pawn {
			rankDiff := int8(m.To.Rank()) - int8(m.From.Rank())
			if rankDiff == 2 || rankDiff == -2 {
				epRank := board.Rank((int8(m.From.Rank()) + int8(m.To.Rank())) / 2)
				result.EnPassant = board.NewSquare(m.From.File(), epRank)
			}
		}
	}

	updateCastlingRights(&result, m, piece, pos)

	result.SideToMove = color.Flip()
	return result
}

func applyCastle(pos *board.Position, m Move, color board.Color) {
	kingSq := m.From
	rookSq := m.To

	king := pos.PieceAt(kingSq)
	rook := pos.PieceAt(rookSq)

	// Determine destination squares based on which side we're castling to.
	var kingDest, rookDest board.Square
	baseRank := kingSq.Rank()

	if rookSq.File() > kingSq.File() {
		// Kingside: king goes to g-file, rook to f-file.
		kingDest = board.NewSquare(board.FileG, baseRank)
		rookDest = board.NewSquare(board.FileF, baseRank)
	} else {
		// Queenside: king goes to c-file, rook to d-file.
		kingDest = board.NewSquare(board.FileC, baseRank)
		rookDest = board.NewSquare(board.FileD, baseRank)
	}

	// Clear both original squares first, then place pieces.
	// This handles the Chess960 case where destination squares might overlap
	// with original squares.
	pos.Remove(kingSq)
	pos.Remove(rookSq)
	pos.Put(king, kingDest)
	pos.Put(rook, rookDest)
}

func applyEnPassantCapture(pos *board.Position, m Move) {
	// Remove the captured pawn. It's on the same file as the destination
	// but on the same rank as the capturing pawn.
	capturedSq := board.NewSquare(m.To.File(), m.From.Rank())
	pos.Remove(capturedSq)
}

func updateCastlingRights(result *board.Position, m Move, piece board.Piece, original *board.Position) {
	color := piece.Color()
	pt := piece.Type()

	// King move: lose all castling rights for this colour.
	if pt == board.King {
		result.Castling.ClearColor(color)
	}

	// Rook move: lose castling right for that rook's file.
	if pt == board.Rook {
		result.Castling.Clear(color, m.From.File())
	}

	// Capture on a corner square: if a rook with castling rights is captured,
	// remove that right. We check if the destination had a rook with castling
	// rights by checking the original position's castling rights for that file.
	toFile := m.To.File()
	toRank := m.To.Rank()

	// Check if a rook on the target square loses castling rights.
	if toRank == board.Rank1 && original.Castling.Has(board.White, toFile) {
		captured := original.PieceAt(m.To)
		if captured == board.NewPiece(board.White, board.Rook) {
			result.Castling.Clear(board.White, toFile)
		}
	}
	if toRank == board.Rank8 && original.Castling.Has(board.Black, toFile) {
		captured := original.PieceAt(m.To)
		if captured == board.NewPiece(board.Black, board.Rook) {
			result.Castling.Clear(board.Black, toFile)
		}
	}
}
