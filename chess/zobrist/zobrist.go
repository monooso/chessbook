package zobrist

import (
	"math/rand/v2"

	"github.com/monooso/chessbook/chess/board"
)

// Table layout:
//   [0..767]   piece-on-square: 12 pieces × 64 squares
//   [768..783] castling rights: 2 colours × 8 files
//   [784..791] en passant file: 8 files
//   [792]      side to move (XORed in when black to move)
const (
	pieceSquareOffset = 0
	castlingOffset    = 768
	enPassantOffset   = 784
	sideToMoveOffset  = 792
	tableSize         = 793
)

// table is the fixed random table, generated deterministically from a fixed seed.
// This seed must never change after the first game is ingested.
var table [tableSize]uint64

func init() {
	// PCG with a fixed seed. The specific seed value is arbitrary but permanent.
	rng := rand.New(rand.NewPCG(0x70B904A5B7D44E02, 0x3F2E517D90C845A1))
	for i := range table {
		table[i] = rng.Uint64()
	}
}

// Hash computes the Zobrist hash of a position.
// En passant is included in the hash only when a legal en passant capture
// exists (see design doc 003, "En passant normalization").
func Hash(pos *board.Position) uint64 {
	var h uint64

	// Piece placement.
	for sq := board.Square(0); sq < 64; sq++ {
		p := pos.PieceAt(sq)
		if !p.IsEmpty() {
			h ^= pieceSquareKey(p, sq)
		}
	}

	// Castling rights.
	for c := board.Color(0); c <= 1; c++ {
		for f := board.FileA; f <= board.FileH; f++ {
			if pos.Castling.Has(c, f) {
				h ^= castlingKey(c, f)
			}
		}
	}

	// En passant (normalized: only if capture is actually possible).
	if pos.EnPassant != board.NoSquare {
		if hasLegalEnPassant(pos) {
			h ^= enPassantKey(pos.EnPassant.File())
		}
	}

	// Side to move.
	if pos.SideToMove == board.Black {
		h ^= table[sideToMoveOffset]
	}

	return h
}

// pieceSquareKey returns the Zobrist key for a piece on a square.
// Piece values are 1–12, mapped to indices 0–11.
func pieceSquareKey(p board.Piece, sq board.Square) uint64 {
	return table[pieceSquareOffset+int(p-1)*64+int(sq)]
}

// castlingKey returns the Zobrist key for a castling right.
func castlingKey(c board.Color, f board.File) uint64 {
	return table[castlingOffset+int(c)*8+int(f)]
}

// enPassantKey returns the Zobrist key for an en passant file.
func enPassantKey(f board.File) uint64 {
	return table[enPassantOffset+int(f)]
}

// hasLegalEnPassant returns true if the en passant capture is actually possible,
// i.e., there is an opposing pawn on an adjacent file that could capture.
func hasLegalEnPassant(pos *board.Position) bool {
	epSq := pos.EnPassant
	epFile := int8(epSq.File())

	// The capturing pawn must be of the side to move.
	capturingColor := pos.SideToMove
	pawn := board.NewPiece(capturingColor, board.Pawn)

	// The capturing pawn is on the rank behind the en passant square
	// (from the capturer's perspective).
	var captureRank board.Rank
	if capturingColor == board.White {
		captureRank = board.Rank(int8(epSq.Rank()) - 1)
	} else {
		captureRank = board.Rank(int8(epSq.Rank()) + 1)
	}

	for _, fd := range []int8{-1, 1} {
		f := epFile + fd
		if f < 0 || f > 7 {
			continue
		}
		sq := board.NewSquare(board.File(f), captureRank)
		if pos.PieceAt(sq) == pawn {
			return true
		}
	}

	return false
}
