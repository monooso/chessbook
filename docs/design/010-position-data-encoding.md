# 010: Position Data Encoding

How the full position state is packed into the `position_data` column (see [008-schema](008-schema.md)). This is the authoritative identity for positions — the Zobrist hash is an index, not a key (see [003-positions-and-identity](003-positions-and-identity.md)).

**Status: tentative.** This encoding is a reasonable starting point, but may change during implementation of the `board` package if a better approach emerges.

## Requirements

The encoding must:

- Be **fixed-size**, so Postgres can index and compare it efficiently via a unique constraint on `BYTEA`
- Capture **piece placement** (which piece on which square), **active colour**, **castling rights**, and **en passant target square** (with normalisation per [003](003-positions-and-identity.md))
- **Exclude** the halfmove clock and fullmove number (see [003](003-positions-and-identity.md))
- Be **deterministic** — the same position always produces the same bytes

## Encoding: nibble-per-square + metadata byte

The board is encoded as 4 bits per square (a nibble), giving 32 bytes for all 64 squares. One additional byte encodes active colour, castling rights, and en passant file.

### Board (32 bytes)

Each nibble encodes the contents of one square. Squares are ordered a1, b1, c1, ..., h8 (rank-major, matching standard board indexing).

| Value | Meaning |
|---|---|
| 0 | Empty |
| 1–6 | White pawn, knight, bishop, rook, queen, king |
| 7–12 | Black pawn, knight, bishop, rook, queen, king |
| 13–15 | Unused |

Two nibbles are packed into each byte, low nibble first (square N in bits 0–3, square N+1 in bits 4–7).

### Metadata (1 byte)

| Bits | Width | Meaning |
|---|---|---|
| 0 | 1 | Active colour (0 = white, 1 = black) |
| 1 | 1 | White kingside castling right |
| 2 | 1 | White queenside castling right |
| 3 | 1 | Black kingside castling right |
| 4 | 1 | Black queenside castling right |
| 5–7 | 3 | En passant file (0 = none, 1–8 = files a–h) |

The en passant file is recorded only when a legal en passant capture exists (see [003](003-positions-and-identity.md), "En passant normalization"). Otherwise it is 0.

### Total: 33 bytes

Fixed-size, deterministic, and compact enough for use as a unique constraint in Postgres. The 3 unused nibble values (13–15) provide room for future extension if needed, though none is anticipated.
