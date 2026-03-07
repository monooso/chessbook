# zobrist package: decisions

## PRNG: PCG with fixed seed

The random table is generated using Go's `math/rand/v2` PCG generator with a fixed seed (`0x70B904A5B7D44E02`, `0x3F2E517D90C845A1`). PCG has good statistical properties for this use case — well-distributed 64-bit values with no practical correlation. The seed values are arbitrary but permanent; changing them would invalidate every hash in the database.

A cryptographic PRNG is not needed — the security of the hashes is irrelevant, only their distribution matters.

## Full recomputation, no incremental updates

`Hash` recomputes from scratch on every call rather than supporting incremental XOR updates. Incremental hashing would be faster during game replay (XOR out old state, XOR in new state per move), but adds complexity and couples the zobrist package to the move package's internals. Full recomputation is O(64) for pieces plus a few extra checks — negligible compared to move generation.

If profiling shows this matters during ingestion, incremental updates can be added without changing the public API.

## En passant normalisation

`hasLegalEnPassant` checks whether an opposing pawn is on an adjacent file in position to capture. This is a simple board state check (two squares at most) and runs only when an en passant square is set. The check ensures that "phantom" en passant squares — recorded in FEN when a pawn advances two squares but no opposing pawn can capture — don't split what should be a single hash bucket.

## Table layout

The table uses a flat array with computed offsets rather than separate arrays for each feature type. This is compact and avoids any indirection. The layout matches design doc 011's specification exactly.
