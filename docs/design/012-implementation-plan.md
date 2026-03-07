# 012: Implementation Plan

The implementation is split into two phases: the chess library (core logic), then the data pipeline (database and ingestion). Each step is built with TDD and validated before the next begins.

## Phase 1: Chess library

Internal Go packages under `chess/`, built in dependency order.

### 1. `board`

Board representation, piece types, square types. The 64-square array and basic operations: get/set piece, empty board, starting position. This is the data structure everything else builds on.

### 2. `fen`

Parse a FEN string into a board state. Serialise a board state back to FEN. Once this works, every subsequent package can set up test positions from FEN strings rather than constructing boards programmatically.

### 3. `move`

Move representation, move generation, legality validation. Covers all piece movement, captures, castling (standard and Chess960), en passant, promotion, pins, and check/checkmate/stalemate detection. Validated with [perft](https://www.chessprogramming.org/Perft) — a standard test suite that counts all legal move paths to a given depth from a given position.

This is the largest and most complex package.

### 4. `zobrist`

Zobrist hashing per [011-zobrist-hashing](011-zobrist-hashing.md). Generate the random table (793 values from a fixed PRNG seed), compute hashes from board state, and support incremental update on move application. Includes en passant normalisation per [003-positions-and-identity](003-positions-and-identity.md).

### 5. Position encoding

The 33-byte binary encoding per [010-position-data-encoding](010-position-data-encoding.md). Serialise board state (piece placement, active colour, castling rights, en passant) to a fixed-size byte sequence and deserialise back. This is the authoritative position identity used for database storage.

Likely lives in the `board` package rather than a separate package, since it's a serialisation of the board's own state.

### 6. `pgn`

Parse PGN files into structured game data: headers (tag pairs) and movetext. Handles the full PGN specification: tags, moves in SAN, comments (discarded), recursive annotation variations (discarded), result markers, and multi-game files.

Does not validate move legality — that is the ingestion layer's responsibility (see Phase 2, step 3). The parser produces a structured representation of what the PGN file contains; the ingestion validator replays the moves and checks them.

### 7. `notation`

UCI to SAN conversion and SAN to UCI conversion, given a board position. Both directions require the board state to resolve ambiguity (see [006-move-notation](006-move-notation.md)).

## Phase 2: Data pipeline

### 1. Database schema

PostgreSQL migrations for the five tables from [008-schema](008-schema.md): `positions`, `edges`, `games`, `position_games`, `edge_games`. All indexes and constraints as specified.

### 2. Ingestion

The two-phase pipeline from [007-ingestion](007-ingestion.md):

- **Validation pass.** Parse every game in the PGN file (using the `pgn` package), then replay each game's moves to verify legality (using the `move` package). Reject the entire file if any game is malformed.
- **Ingestion pass.** For each validated game: insert the game record, replay the moves, upsert positions and edges, insert join table rows. Each game is an atomic transaction. Duplicate detection via hash of Seven Tag Roster + movetext.

### 3. Query layer

The core read path from [008-schema](008-schema.md): look up a position by Zobrist hash + position data, fetch outgoing edges, filter game IDs against game metadata, compute win/loss/draw counts. This is the interface that a future UI or API would call.
