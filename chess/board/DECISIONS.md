# board package: decisions

## Piece encoding

Pieces are encoded as a single `int8`: White pieces 1–6, Black pieces 7–12, NoPiece = 0. This directly matches the nibble encoding in design doc 010, so serialising a position to the 33-byte binary format is a straightforward copy of values — no translation layer needed.

The `Color()` and `Type()` methods derive colour and piece type arithmetically (`(p-1)/6` and `(p-1)%6+1`) rather than using a lookup table. This is compact and avoids maintaining a separate mapping, at the cost of being slightly less obvious to read. The test coverage for all 12 piece values gives confidence in the arithmetic.

## CastlingRights as [2]uint8

Castling rights are a `[2]uint8` array indexed by `Color`, with each bit representing a file (bit 0 = a-file through bit 7 = h-file). This was initially a struct with `White` and `Black` fields, but the array eliminates if/else branching in every method since `Color` values are explicitly 0 and 1.

The per-file bitmask supports Chess960 (where rooks can start on any file) and maps directly to the 16 Zobrist castling keys in design doc 011 — one key per (colour, file) combination.

## Position excludes halfmove clock and fullmove number

The `Position` type does not include the halfmove clock or fullmove number. These are excluded from position identity (design doc 003), and are not needed for move legality validation — the fifty-move rule is a draw claim, not a constraint on legal moves. The `fen` package will need to parse these fields (they're part of the FEN spec) but can discard them or handle them separately.

## Square ordering

Squares use file-major ordering: a1=0, b1=1, ..., h1=7, a2=8, ..., h8=63. This is the Little-Endian Rank-File (LERF) mapping, which is the most common convention in chess programming. It makes file and rank extraction simple modular arithmetic.
