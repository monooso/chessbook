package board

// EncodedSize is the size of an encoded position in bytes.
// 32 bytes board (nibble per square) + 2 bytes castling + 1 byte metadata.
const EncodedSize = 35

// EncodedPosition is the fixed-size binary encoding of a position,
// used as the authoritative identity in the database (see design doc 010).
type EncodedPosition = [EncodedSize]byte

// Encode serialises a position to a fixed-size binary representation.
// En passant is normalised: the EP square is only included when a legal
// en passant capture exists (matching the Zobrist normalisation in design doc 003).
func Encode(pos *Position) EncodedPosition {
	var buf EncodedPosition

	// Board: 32 bytes, nibble per square, low nibble first.
	for i := 0; i < 64; i += 2 {
		lo := byte(pos.Squares[i])
		hi := byte(pos.Squares[i+1])
		buf[i/2] = lo | (hi << 4)
	}

	// Castling: 2 bytes (white, black), directly from CastlingRights.
	buf[32] = pos.Castling[White]
	buf[33] = pos.Castling[Black]

	// Metadata: 1 byte.
	// Bit 0: active colour (0=white, 1=black).
	// Bits 1–4: en passant file (0=none, 1–8=files a–h).
	var meta byte
	if pos.SideToMove == Black {
		meta |= 1
	}
	if pos.EnPassant != NoSquare && hasLegalEnPassant(pos) {
		meta |= byte(pos.EnPassant.File()+1) << 1
	}
	buf[34] = meta

	return buf
}

// hasLegalEnPassant returns true if an en passant capture is actually possible,
// i.e., there is an opposing pawn on an adjacent file that could capture.
func hasLegalEnPassant(pos *Position) bool {
	epFile := int8(pos.EnPassant.File())

	// The capturing pawn must be of the side to move.
	pawn := NewPiece(pos.SideToMove, Pawn)

	// The capturing pawn is on the rank behind the en passant square.
	var captureRank Rank
	if pos.SideToMove == White {
		captureRank = Rank(int8(pos.EnPassant.Rank()) - 1)
	} else {
		captureRank = Rank(int8(pos.EnPassant.Rank()) + 1)
	}

	for _, fd := range []int8{-1, 1} {
		f := epFile + fd
		if f < 0 || f > 7 {
			continue
		}
		sq := NewSquare(File(f), captureRank)
		if pos.PieceAt(sq) == pawn {
			return true
		}
	}

	return false
}

// Decode deserialises a fixed-size binary representation back to a Position.
func Decode(buf EncodedPosition) Position {
	var pos Position

	// Board: 32 bytes, nibble per square.
	for i := 0; i < 64; i += 2 {
		b := buf[i/2]
		pos.Squares[i] = Piece(b & 0x0F)
		pos.Squares[i+1] = Piece(b >> 4)
	}

	// Castling: 2 bytes.
	pos.Castling[White] = buf[32]
	pos.Castling[Black] = buf[33]

	// Metadata.
	meta := buf[34]
	if meta&1 != 0 {
		pos.SideToMove = Black
	}
	epFile := (meta >> 1) & 0x0F
	if epFile > 0 {
		// En passant file is stored as 1-8, map to 0-7.
		// The rank is determined by the side to move: rank 6 for white
		// (black just double-pushed), rank 3 for black (white just double-pushed).
		var epRank Rank
		if pos.SideToMove == White {
			epRank = Rank6
		} else {
			epRank = Rank3
		}
		pos.EnPassant = NewSquare(File(epFile-1), epRank)
	} else {
		pos.EnPassant = NoSquare
	}

	return pos
}
