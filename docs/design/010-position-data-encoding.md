# 010: Position Data Encoding

How the full position state is packed into the `position_data` column (see [008-schema](008-schema.md)). This is the authoritative identity for positions — the Zobrist hash is an index, not a key (see [003-positions-and-identity](003-positions-and-identity.md)).

## Requirements

The encoding must:

- Be **fixed-size**, so Postgres can index and compare it efficiently via a unique constraint on `BYTEA`
- Capture **piece placement** (which piece on which square), **active colour**, **castling rights**, and **en passant target square** (with normalisation per [003](003-positions-and-identity.md))
- **Exclude** the halfmove clock and fullmove number (see [003](003-positions-and-identity.md))
- Be **deterministic** — the same position always produces the same bytes
- Support **Chess960 castling rights** (rooks on arbitrary files, not just a- and h-files)

## Encoding: nibble-per-square + castling + metadata

The board is encoded as 4 bits per square (a nibble), giving 32 bytes for all 64 squares. Two additional bytes encode castling rights (one per colour), and one byte encodes active colour and en passant file.

### Board (32 bytes)

Each nibble encodes the contents of one square. Squares are ordered a1, b1, c1, ..., h8 (rank-major, matching standard board indexing).

| Value | Meaning |
|---|---|
| 0 | Empty |
| 1–6 | White pawn, knight, bishop, rook, queen, king |
| 7–12 | Black pawn, knight, bishop, rook, queen, king |
| 13–15 | Unused |

Two nibbles are packed into each byte, low nibble first (square N in bits 0–3, square N+1 in bits 4–7).

### Castling rights (2 bytes)

One byte per colour (white first, then black). Each byte is a bitmask where bit N corresponds to file N (bit 0 = a-file, bit 7 = h-file). This directly matches the `CastlingRights [2]uint8` representation in the board package and the 16 Zobrist castling keys in [011-zobrist-hashing](011-zobrist-hashing.md).

For standard chess, the bytes will typically be `0b10000001` (a- and h-files) or subsets thereof. For Chess960, any combination of files is possible.

### Metadata (1 byte)

| Bits | Width | Meaning |
|---|---|---|
| 0 | 1 | Active colour (0 = white, 1 = black) |
| 1–4 | 4 | En passant file (0 = none, 1–8 = files a–h) |
| 5–7 | 3 | Unused |

The en passant file is recorded only when a legal en passant capture exists (see [003](003-positions-and-identity.md), "En passant normalization"). Otherwise it is 0.

### Total: 35 bytes

Fixed-size, deterministic, and compact enough for use as a unique constraint in Postgres. The original design specified 33 bytes using 4-bit castling flags (KQkq), but this was expanded to 35 bytes during implementation to support Chess960 castling rights on arbitrary files.
