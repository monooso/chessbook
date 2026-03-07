# 009: Technology Choices

## Language: Go

Go is the implementation language. The key factors:

- **PostgreSQL ecosystem.** `pgx` provides a mature driver with COPY and batch support, which matters for ingestion at scale (tens of millions of games).
- **Concurrency.** Goroutines and channels map naturally to the file-level parallel ingestion described in [007-ingestion](007-ingestion.md).
- **Single binary.** Simple deployment and flexible reuse of the final product.
- **Suitable for the workload.** This is a data-heavy backend with no UI requirements in scope. Go's performance characteristics and standard library are well-suited.

## Chess library: custom

The chess logic is implemented as internal packages within this project, not pulled from a third-party library.

### Why

- **Learning.** This is in part a learning project. The chess domain (board representation, move generation, FEN, PGN, Zobrist hashing) is well-specified with extensive prior art, making it a tractable and rewarding implementation target.
- **Control over position identity.** The design requires specific semantics for position keys: en passant normalisation (see [003-positions-and-identity](003-positions-and-identity.md)), halfmove clock exclusion, and a compact binary encoding for `position_data`. Owning the implementation avoids working around a library's assumptions.
- **The Go ecosystem isn't compelling here.** The most popular chess library (`notnil/chess`) is archived. The alternatives are small projects. There isn't a dominant, well-maintained option that would make rolling our own clearly wasteful.
- **Bounded scope.** We need board representation, move generation/validation, FEN parsing, PGN parsing, UCI/SAN conversion, and Zobrist hashing. We do not need a chess engine (no search, no evaluation). This is the "data structures and rules" layer.

### Correctness

Move generation — the most complex part — is validated using perft (performance test, move path enumeration), a standard test suite in chess programming that counts all legal move paths to a given depth from a given position. This provides high confidence in correctness across edge cases (pins, en passant, castling through check, promotion).

### Structure

The chess logic is organised as separate Go packages within a single module. Clear boundaries and independent testability, but no pretence of independent distribution. Interfaces can evolve freely as the implementation matures.

```
chessbook/
  chess/
    board/      # board representation, piece types, squares
    move/       # move generation, validation
    fen/        # FEN parsing/serialisation
    zobrist/    # Zobrist hashing
    pgn/        # PGN parsing
    notation/   # UCI/SAN conversion
```

If any package later proves useful as a standalone library, extraction is straightforward. But the distribution tax (stable public APIs, semantic versioning, documentation for external consumers) is not paid upfront.

### Build order

The packages are built in dependency order, each testable before the next begins:

1. `board` — board representation, piece types, squares
2. `fen` — FEN parsing and serialisation
3. `move` — move generation, validation (validated with perft)
4. `zobrist` — Zobrist hashing with en passant normalisation
5. `pgn` — PGN parsing
6. `notation` — UCI/SAN conversion
