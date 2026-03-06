# 006: Move Notation

Moves are stored in UCI (Universal Chess Interface) notation and converted to SAN (Standard Algebraic Notation) at display time.

## UCI for storage

A UCI move is the source square concatenated with the destination square, plus an optional promotion piece: `g1f3`, `e2e4`, `e7e8q`. This is always 4 or 5 characters, unambiguous without context, and trivially parseable.

SAN (`Nf3`, `e4`, `Qxd7+`) is more readable to a chess player but is position-dependent. The same move can have different SAN representations depending on the board state — whether disambiguation is needed, whether a capture occurred, and whether the move gives check or checkmate. SAN is a presentation concern, not a storage concern.

## Deriving SAN at display time

Converting UCI to SAN requires the board position. The steps are:

1. **Look up the piece on the source square** — one array index into the board representation.
2. **Check for disambiguation** — determine whether another piece of the same type can also reach the destination square. This is the only non-trivial step, and it applies only when multiple same-type pieces exist (at most 2 knights, 2 rooks, rarely more from promotion). In practice this check is needed infrequently.
3. **Check for capture** — one array index to see if the destination square is occupied.
4. **Check for check/checkmate** — one move-generation pass on the resulting position.

Steps 1 and 3 are effectively free. Step 2 is rare. Step 4 is the most expensive, but single-position move generation is sub-microsecond in any reasonable chess library.

## Why the cost is negligible

SAN derivation only runs for moves that are actually displayed — at most a few dozen at a time (the moves from the current position, possibly a variation line). Compared to the operations already in the pipeline — filtering across thousands of positions, computing win/loss/draw deltas, traversing variations — the cost of SAN derivation is noise.

The concern would be different if we were batch-converting entire databases on load, but we are not. The conversion happens on the display path, for the small set of moves visible on screen.

## Conversions in both directions require position

Neither direction — UCI to SAN or SAN to UCI — is a pure string transformation. Both require the board position:

- **UCI → SAN**: need the position to determine the piece type, disambiguation, capture status, and check/checkmate.
- **SAN → UCI**: need the position to determine which piece on which square the notation refers to.

This is fine. During ingestion the position is available because we are replaying the game move by move to build the graph. During display the position is available because we need it to render the board. There is no point in the pipeline where we need to convert notation without access to the position.
