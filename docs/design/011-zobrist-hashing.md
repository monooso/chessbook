# 011: Zobrist Hashing

How Zobrist hash keys are generated and applied. See [003-positions-and-identity](003-positions-and-identity.md) for how hashes fit into the position identity model.

## Background

A Zobrist hash assigns a random 64-bit value to each feature of a chess position (piece on square, castling right, en passant file, side to move) and XORs them together to produce a single integer key. The hash can be updated incrementally as moves are made — XOR out the old state, XOR in the new state — making it efficient for replaying games move by move during ingestion.

For a general introduction, see the [Chessprogramming wiki on Zobrist Hashing](https://www.chessprogramming.org/Zobrist_Hashing) and the [Wikipedia article](https://en.wikipedia.org/wiki/Zobrist_hashing).

## Random table: custom, not Polyglot

The [Polyglot opening book format](http://hgm.nubati.net/book_format.html) defines a widely-used set of 781 random values for Zobrist hashing. We do not use these values, because Polyglot cannot fully represent Chess960 castling rights.

### The Chess960 castling problem

In standard chess, castling rights are one of four flags: white kingside, white queenside, black kingside, black queenside. The king always starts on e1/e8 and the rooks on a1/h1/a8/h8, so these four flags are unambiguous. Polyglot assigns one random value to each, for a total of 4 castling keys.

In Chess960, the king and rooks start on arbitrary squares. Castling rights are tied to specific rooks, identified by file. A position might have castling rights for a rook on the b-file rather than the a-file, and these are different rights that must produce different hashes. This requires up to 16 castling keys (files a–h for each colour).

The Polyglot community discussed a backwards-compatible extension — using the standard 4 keys for outermost-rook castling and adding 12 new keys for inner files — but no standard values were ever agreed. The relevant discussions:

- [TalkChess: Values for Polyglot c960 zobrist hashing?](https://talkchess.com/viewtopic.php?t=76957)
- [TalkChess: Polyglot books and Chess960](https://talkchess.com/viewtopic.php?t=25654)

XBoard works around this by forbidding non-outermost-rook castling rights in opening books. The python-chess library added an undocumented `ZobristHasher` class to allow custom extensions but does not define standard Chess960 values.

Since our design explicitly supports Chess960 (see [003-positions-and-identity](003-positions-and-identity.md) and [006-move-notation](006-move-notation.md)), we need castling keys that cover all file positions. Polyglot cannot provide this, so we generate our own random table.

## Random table specification

The random table is generated deterministically from a fixed PRNG seed, so that hashes are stable across builds and runs. The seed and generation algorithm are part of the specification — changing either would invalidate every hash in the database.

### Values required

| Feature | Count | Notes |
|---|---|---|
| Piece on square | 768 | 12 piece types (6 pieces × 2 colours) × 64 squares |
| Castling rights | 16 | Files a–h for each colour |
| En passant file | 8 | Files a–h |
| Side to move | 1 | XORed in when it is black's turn |
| **Total** | **793** | |

### En passant normalisation

The en passant value is XORed into the hash only when a legal en passant capture exists — i.e., an opposing pawn is on an adjacent file in position to capture (see [003-positions-and-identity](003-positions-and-identity.md), "En passant normalization"). This matches the approach used by Polyglot and most chess engines, and avoids splitting functionally identical positions into distinct hash buckets.

### PRNG and seed

The specific PRNG algorithm and seed value are implementation decisions, to be fixed when the `zobrist` package is built. The requirements are:

- **Deterministic.** Same seed always produces the same sequence.
- **Good distribution.** Values should be well-spread across the 64-bit space to minimise hash collisions. A cryptographic PRNG is not required, but a low-quality generator (e.g. simple LCG) is not sufficient.
- **Documented and pinned.** The algorithm and seed are recorded in the source code and must not change after the first game is ingested.

## Cross-checking

Our hashes will not match Polyglot hashes (different random values), but correctness can still be validated by:

- Verifying that two positions considered identical by our system have the same hash
- Verifying that two positions with different legal moves have different hashes (within collision probability)
- Comparing the set of positions our system considers distinct against a Polyglot implementation for standard chess positions — the sets should agree
- Perft-based testing of the underlying board representation, which the hash depends on
