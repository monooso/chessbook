package notation

import (
	"fmt"
	"strings"

	"github.com/monooso/chessbook/chess/board"
	"github.com/monooso/chessbook/chess/move"
)

// pieceTypeFromSAN maps SAN piece letters to piece types.
var pieceTypeFromSAN = map[byte]board.PieceType{
	'N': board.Knight,
	'B': board.Bishop,
	'R': board.Rook,
	'Q': board.Queen,
	'K': board.King,
}

// sanFromPieceType maps piece types to SAN letters.
var sanFromPieceType = [7]byte{0, 0, 'N', 'B', 'R', 'Q', 'K'}

// promotionFromSAN maps SAN promotion piece letters to piece types.
var promotionFromSAN = map[byte]board.PieceType{
	'Q': board.Queen,
	'R': board.Rook,
	'B': board.Bishop,
	'N': board.Knight,
}

// SANToMove converts a SAN string to a Move, given the current position.
// The SAN is matched against legal moves generated from the position.
func SANToMove(pos *board.Position, san string) (move.Move, error) {
	if len(san) == 0 {
		return move.Move{}, fmt.Errorf("notation: empty SAN string")
	}

	// Strip check/checkmate markers.
	san = strings.TrimRight(san, "+#")

	// Castling.
	if san == "O-O" || san == "O-O-O" {
		return matchCastling(pos, san)
	}

	// Parse the SAN components.
	parsed, err := parseSAN(san)
	if err != nil {
		return move.Move{}, err
	}

	// Find the matching legal move.
	moves := move.Generate(pos)
	var matches []move.Move

	for _, m := range moves {
		if m.Castle {
			continue
		}

		piece := pos.PieceAt(m.From)
		if piece.Type() != parsed.pieceType {
			continue
		}
		if m.To != parsed.toSquare {
			continue
		}
		if parsed.promotion != 0 && m.Promotion != parsed.promotion {
			continue
		}
		if parsed.promotion == 0 && m.Promotion != 0 {
			continue
		}

		// Check disambiguation.
		if parsed.fromFile >= 0 && board.File(parsed.fromFile) != m.From.File() {
			continue
		}
		if parsed.fromRank >= 0 && board.Rank(parsed.fromRank) != m.From.Rank() {
			continue
		}

		matches = append(matches, m)
	}

	switch len(matches) {
	case 0:
		return move.Move{}, fmt.Errorf("notation: no legal move matches %q", san)
	case 1:
		return matches[0], nil
	default:
		return move.Move{}, fmt.Errorf("notation: ambiguous SAN %q matches %d moves", san, len(matches))
	}
}

type parsedSAN struct {
	pieceType board.PieceType
	toSquare  board.Square
	fromFile  int8 // -1 if not specified
	fromRank  int8 // -1 if not specified
	promotion board.PieceType
}

func parseSAN(san string) (parsedSAN, error) {
	p := parsedSAN{fromFile: -1, fromRank: -1}
	i := 0

	// Piece type (uppercase letter) or pawn (lowercase).
	if i < len(san) && san[i] >= 'A' && san[i] <= 'Z' {
		pt, ok := pieceTypeFromSAN[san[i]]
		if !ok {
			return p, fmt.Errorf("notation: invalid piece %q in SAN", san[i])
		}
		p.pieceType = pt
		i++
	} else {
		p.pieceType = board.Pawn
	}

	// Collect remaining characters, stripping 'x' for captures.
	rest := san[i:]
	rest = strings.ReplaceAll(rest, "x", "")

	// For pawns, the format is: [file]file+rank[=promotion]
	// For pieces, the format is: [file][rank]file+rank
	// We need to find the target square (last two chars before promotion).

	// Check for promotion.
	if eqIdx := strings.IndexByte(rest, '='); eqIdx >= 0 {
		if eqIdx+1 >= len(rest) {
			return p, fmt.Errorf("notation: missing promotion piece in %q", san)
		}
		pt, ok := promotionFromSAN[rest[eqIdx+1]]
		if !ok {
			return p, fmt.Errorf("notation: invalid promotion piece in %q", san)
		}
		p.promotion = pt
		rest = rest[:eqIdx]
	}

	// The last two characters of rest should be the target square.
	if len(rest) < 2 {
		return p, fmt.Errorf("notation: cannot parse target square from %q", san)
	}

	toFile := rest[len(rest)-2]
	toRank := rest[len(rest)-1]
	if toFile < 'a' || toFile > 'h' || toRank < '1' || toRank > '8' {
		return p, fmt.Errorf("notation: invalid target square in %q", san)
	}
	p.toSquare = board.NewSquare(board.File(toFile-'a'), board.Rank(toRank-'1'))

	// Anything before the target square is disambiguation.
	disambig := rest[:len(rest)-2]
	for _, ch := range []byte(disambig) {
		if ch >= 'a' && ch <= 'h' {
			p.fromFile = int8(ch - 'a')
		} else if ch >= '1' && ch <= '8' {
			p.fromRank = int8(ch - '1')
		}
	}

	return p, nil
}

func matchCastling(pos *board.Position, san string) (move.Move, error) {
	moves := move.Generate(pos)

	kingSq := board.FindKing(pos, pos.SideToMove)
	if kingSq == board.NoSquare {
		return move.Move{}, fmt.Errorf("notation: no king found for castling")
	}

	for _, m := range moves {
		if !m.Castle {
			continue
		}
		rookFile := m.To.File()
		kingFile := kingSq.File()
		isKingside := rookFile > kingFile

		if san == "O-O" && isKingside {
			return m, nil
		}
		if san == "O-O-O" && !isKingside {
			return m, nil
		}
	}

	return move.Move{}, fmt.Errorf("notation: castling %q not legal", san)
}

// MoveToSAN converts a Move to SAN notation, given the current position.
func MoveToSAN(pos *board.Position, m move.Move) string {
	// Castling.
	if m.Castle {
		if m.To.File() > m.From.File() {
			return addCheckSuffix(pos, m, "O-O")
		}
		return addCheckSuffix(pos, m, "O-O-O")
	}

	piece := pos.PieceAt(m.From)
	pt := piece.Type()
	var b strings.Builder

	if pt == board.Pawn {
		// Pawn moves: file if capture, then target square, then promotion.
		isCapture := !pos.PieceAt(m.To).IsEmpty() || m.To == pos.EnPassant
		if isCapture {
			b.WriteByte(byte('a' + m.From.File()))
			b.WriteByte('x')
		}
		b.WriteString(m.To.String())
		if m.Promotion != 0 {
			b.WriteByte('=')
			b.WriteByte(sanFromPieceType[m.Promotion])
		}
	} else {
		// Piece moves: piece letter, disambiguation, capture, target square.
		b.WriteByte(sanFromPieceType[pt])

		// Disambiguation: check if another piece of the same type can reach the same square.
		disambig := disambiguation(pos, m, pt)
		b.WriteString(disambig)

		if !pos.PieceAt(m.To).IsEmpty() {
			b.WriteByte('x')
		}
		b.WriteString(m.To.String())
	}

	return addCheckSuffix(pos, m, b.String())
}

// disambiguation returns the disambiguation string needed for a piece move.
func disambiguation(pos *board.Position, m move.Move, pt board.PieceType) string {
	color := pos.SideToMove
	moves := move.Generate(pos)

	needFile := false
	needRank := false

	for _, other := range moves {
		if other.Castle || other.From == m.From || other.To != m.To {
			continue
		}
		otherPiece := pos.PieceAt(other.From)
		if otherPiece.Type() != pt || otherPiece.Color() != color {
			continue
		}

		// Another piece of the same type can reach the same square.
		if other.From.File() != m.From.File() {
			needFile = true
		} else if other.From.Rank() != m.From.Rank() {
			needRank = true
		} else {
			// Same file and rank — shouldn't happen, but include both.
			needFile = true
			needRank = true
		}
	}

	if needFile && needRank {
		return m.From.String()
	} else if needFile {
		return string(rune('a' + m.From.File()))
	} else if needRank {
		return string(rune('1' + m.From.Rank()))
	}
	return ""
}

func addCheckSuffix(pos *board.Position, m move.Move, san string) string {
	after := move.Apply(pos, m)
	if !move.IsInCheck(&after) {
		return san
	}

	// Check if it's checkmate (no legal moves).
	if len(move.Generate(&after)) == 0 {
		return san + "#"
	}
	return san + "+"
}
