package board

// EncodedSize is the size of an encoded position in bytes.
// 32 bytes board (nibble per square) + 2 bytes castling + 1 byte metadata.
const EncodedSize = 35

// EncodedPosition is the fixed-size binary encoding of a position,
// used as the authoritative identity in the database (see design doc 010).
type EncodedPosition = [EncodedSize]byte

// Encode serialises a position to a fixed-size binary representation.
// The encoding does NOT perform en passant normalisation — that is the
// caller's responsibility. The zobrist and ingestion layers handle
// normalisation before encoding.
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
	if pos.EnPassant != NoSquare {
		meta |= byte(pos.EnPassant.File()+1) << 1
	}
	buf[34] = meta

	return buf
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
