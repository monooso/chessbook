# fen package: decisions

## Only 4 fields output

`Format` outputs only the first 4 FEN fields (piece placement, active colour, castling, en passant). The halfmove clock and fullmove number are excluded from position identity (design doc 003), and the `Position` type doesn't store them. `Parse` accepts but discards fields 5 and 6 if present, and requires at least 4 fields.

## Castling notation: hybrid standard + Shredder-FEN

**Parsing** accepts both notations:
- Standard `KQkq` (maps K→h-file, Q→a-file)
- Shredder-FEN file letters `A`–`H` / `a`–`h` (maps directly to file index)

This means standard FEN and Chess960 FEN both parse correctly without the caller needing to specify which format to expect.

**Formatting** uses a hybrid approach: `K`/`Q`/`k`/`q` for a-file and h-file rook positions (standard chess), and file letters for non-standard positions (Chess960). Output order is: kingside (h-file) first, then non-standard files alphabetically, then queenside (a-file). This produces standard `KQkq` output for standard positions and unambiguous Shredder-FEN for Chess960.

## Lenient castling parsing

Duplicate characters in the castling field (e.g. "KK") are accepted without error — setting the same bit twice is idempotent. Strictly this is invalid FEN, but rejecting it would add complexity for no practical benefit. The ingestion layer's file validation phase is where strict correctness checks belong; the FEN parser is lenient on inputs that don't produce incorrect state.
