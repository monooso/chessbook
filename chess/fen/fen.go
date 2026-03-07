package fen

import (
	"fmt"
	"strings"

	"github.com/monooso/chessbook/chess/board"
)

// pieceFromChar maps FEN piece characters to Piece values.
var pieceFromChar = map[byte]board.Piece{
	'P': board.NewPiece(board.White, board.Pawn),
	'N': board.NewPiece(board.White, board.Knight),
	'B': board.NewPiece(board.White, board.Bishop),
	'R': board.NewPiece(board.White, board.Rook),
	'Q': board.NewPiece(board.White, board.Queen),
	'K': board.NewPiece(board.White, board.King),
	'p': board.NewPiece(board.Black, board.Pawn),
	'n': board.NewPiece(board.Black, board.Knight),
	'b': board.NewPiece(board.Black, board.Bishop),
	'r': board.NewPiece(board.Black, board.Rook),
	'q': board.NewPiece(board.Black, board.Queen),
	'k': board.NewPiece(board.Black, board.King),
}

// charFromPiece maps Piece values to FEN characters.
var charFromPiece [13]byte

func init() {
	for ch, p := range pieceFromChar {
		charFromPiece[p] = ch
	}
}

// Parse parses a FEN string into a Position.
// The first 4 fields are required (piece placement, active colour, castling, en passant).
// Fields 5 and 6 (halfmove clock, fullmove number) are accepted but ignored,
// as they are excluded from position identity (see design doc 003).
func Parse(s string) (board.Position, error) {
	fields := strings.Fields(s)
	if len(fields) < 4 {
		return board.Position{}, fmt.Errorf("fen: expected at least 4 fields, got %d", len(fields))
	}

	pos := board.EmptyPosition()

	if err := parsePieces(&pos, fields[0]); err != nil {
		return board.Position{}, err
	}
	if err := parseActiveColor(&pos, fields[1]); err != nil {
		return board.Position{}, err
	}
	if err := parseCastling(&pos, fields[2]); err != nil {
		return board.Position{}, err
	}
	if err := parseEnPassant(&pos, fields[3]); err != nil {
		return board.Position{}, err
	}

	return pos, nil
}

func parsePieces(pos *board.Position, s string) error {
	ranks := strings.Split(s, "/")
	if len(ranks) != 8 {
		return fmt.Errorf("fen: expected 8 ranks, got %d", len(ranks))
	}

	// FEN ranks are ordered 8 to 1 (top to bottom).
	for i, rankStr := range ranks {
		rank := board.Rank(7 - i)
		file := board.FileA

		for j := 0; j < len(rankStr); j++ {
			ch := rankStr[j]
			if ch >= '1' && ch <= '8' {
				file += board.File(ch - '0')
			} else if p, ok := pieceFromChar[ch]; ok {
				if file > board.FileH {
					return fmt.Errorf("fen: too many squares in rank %d", 8-i)
				}
				pos.Put(p, board.NewSquare(file, rank))
				file++
			} else {
				return fmt.Errorf("fen: invalid piece character %q", ch)
			}
		}

		if file != board.FileH+1 {
			return fmt.Errorf("fen: rank %d has %d squares, want 8", 8-i, file)
		}
	}

	return nil
}

func parseActiveColor(pos *board.Position, s string) error {
	switch s {
	case "w":
		pos.SideToMove = board.White
	case "b":
		pos.SideToMove = board.Black
	default:
		return fmt.Errorf("fen: invalid active color %q", s)
	}
	return nil
}

func parseCastling(pos *board.Position, s string) error {
	if s == "-" {
		return nil
	}

	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch == 'K':
			pos.Castling.Set(board.White, board.FileH)
		case ch == 'Q':
			pos.Castling.Set(board.White, board.FileA)
		case ch == 'k':
			pos.Castling.Set(board.Black, board.FileH)
		case ch == 'q':
			pos.Castling.Set(board.Black, board.FileA)
		case ch >= 'A' && ch <= 'H':
			pos.Castling.Set(board.White, board.File(ch-'A'))
		case ch >= 'a' && ch <= 'h':
			pos.Castling.Set(board.Black, board.File(ch-'a'))
		default:
			return fmt.Errorf("fen: invalid castling character %q", ch)
		}
	}

	return nil
}

func parseEnPassant(pos *board.Position, s string) error {
	if s == "-" {
		return nil
	}

	if len(s) != 2 {
		return fmt.Errorf("fen: invalid en passant square %q", s)
	}

	file := s[0]
	rank := s[1]

	if file < 'a' || file > 'h' || rank < '1' || rank > '8' {
		return fmt.Errorf("fen: invalid en passant square %q", s)
	}

	pos.EnPassant = board.NewSquare(board.File(file-'a'), board.Rank(rank-'1'))
	return nil
}

// Format serialises a Position to a FEN string (first 4 fields only).
// Standard castling notation (KQkq) is used when rooks are on the a- and h-files.
// Shredder-FEN notation (file letters) is used for other rook positions (Chess960).
func Format(pos board.Position) string {
	var b strings.Builder

	formatPieces(&b, &pos)
	b.WriteByte(' ')
	formatActiveColor(&b, pos.SideToMove)
	b.WriteByte(' ')
	formatCastling(&b, pos.Castling)
	b.WriteByte(' ')
	formatEnPassant(&b, pos.EnPassant)

	return b.String()
}

func formatPieces(b *strings.Builder, pos *board.Position) {
	for rank := board.Rank8; rank >= board.Rank1; rank-- {
		if rank < board.Rank8 {
			b.WriteByte('/')
		}

		empty := 0
		for file := board.FileA; file <= board.FileH; file++ {
			p := pos.PieceAt(board.NewSquare(file, rank))
			if p.IsEmpty() {
				empty++
				continue
			}
			if empty > 0 {
				b.WriteByte(byte('0' + empty))
				empty = 0
			}
			b.WriteByte(charFromPiece[p])
		}
		if empty > 0 {
			b.WriteByte(byte('0' + empty))
		}
	}
}

func formatActiveColor(b *strings.Builder, c board.Color) {
	if c == board.White {
		b.WriteByte('w')
	} else {
		b.WriteByte('b')
	}
}

func formatCastling(b *strings.Builder, cr board.CastlingRights) {
	if cr == [2]uint8{} {
		b.WriteByte('-')
		return
	}

	// For each colour, output castling rights in standard order:
	// K/k before Q/q for standard rook positions (h- and a-files),
	// then file letters (A–H / a–h) alphabetically for non-standard positions (Chess960).
	for _, entry := range []struct {
		color    board.Color
		stdKing  byte // Standard notation for kingside (h-file)
		stdQueen byte // Standard notation for queenside (a-file)
		fileBase byte // Base character for Shredder-FEN file letters
	}{
		{board.White, 'K', 'Q', 'A'},
		{board.Black, 'k', 'q', 'a'},
	} {
		// Kingside (h-file) first, using standard notation.
		if cr.Has(entry.color, board.FileH) {
			b.WriteByte(entry.stdKing)
		}
		// Non-standard files in alphabetical order.
		for file := board.FileB; file <= board.FileG; file++ {
			if cr.Has(entry.color, file) {
				b.WriteByte(entry.fileBase + byte(file))
			}
		}
		// Queenside (a-file) last, using standard notation.
		if cr.Has(entry.color, board.FileA) {
			b.WriteByte(entry.stdQueen)
		}
	}
}

func formatEnPassant(b *strings.Builder, sq board.Square) {
	if sq == board.NoSquare {
		b.WriteByte('-')
		return
	}
	b.WriteString(sq.String())
}
