# Chessbook

A chess position graph database. Ingests PGN files and builds a directed graph of positions (nodes) connected by moves (edges), with game IDs on both. Statistics (win/loss/draw) are computed at query time by filtering game IDs against game metadata.

## Prerequisites

- **Go 1.26+** (managed via [mise](https://mise.jdx.dev/))
- **PostgreSQL 18+** (running on `localhost:5432`)

## Quick start

```bash
# Install dependencies
mise install
go mod download

# Create the database
createdb chessbook
createdb chessbook_test  # for tests

# Run migrations
# (programmatically via db.Connect + db.Migrate — see Usage below)

# Run tests
go test ./... -count=1 -p 1
```

The `-p 1` flag is required because the `ingest` and `query` packages share a test database and must not run in parallel.

## Project structure

```
chess/                  # Chess logic (no database dependency)
  board/                # Squares, pieces, positions, encoding
  fen/                  # FEN parsing and formatting
  move/                 # Move generation, application, attack detection
  zobrist/              # Zobrist hashing
  pgn/                  # PGN file parsing
  notation/             # SAN ↔ Move conversion

db/                     # Database connection and migrations
  migrations/           # SQL migration files

ingest/                 # PGN ingestion pipeline
query/                  # Position lookup and statistics

docs/design/            # Architecture and design documents (001–012)
```

## Usage

### Connecting and migrating

```go
import (
    "context"
    "github.com/monooso/chessbook/db"
)

ctx := context.Background()
pool, err := db.Connect(ctx, "postgres://user:pass@localhost:5432/chessbook?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

if err := db.Migrate(ctx, pool); err != nil {
    log.Fatal(err)
}
```

### Ingesting a PGN file

Ingestion is two-phase: validate first (reject entire file if any game is malformed), then ingest game-by-game (each game is an atomic transaction).

```go
import (
    "os"
    "github.com/monooso/chessbook/chess/pgn"
    "github.com/monooso/chessbook/ingest"
)

f, _ := os.Open("games.pgn")
defer f.Close()

// Phase 1: Parse and validate.
games, err := pgn.Parse(f)
if err != nil {
    log.Fatal(err)
}

validated, err := ingest.ValidateFile(games)
if err != nil {
    log.Fatal(err) // Entire file rejected.
}

// Phase 2: Ingest into database.
result, err := ingest.IngestFile(ctx, pool, validated, games)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Ingested: %d, Duplicates: %d, Errors: %d\n",
    result.Ingested, result.Duplicates, len(result.Errors))
```

### Querying a position

```go
import (
    "github.com/monooso/chessbook/chess/board"
    "github.com/monooso/chessbook/chess/fen"
    "github.com/monooso/chessbook/chess/zobrist"
    "github.com/monooso/chessbook/query"
)

// Look up the starting position.
pos, _ := fen.Parse("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq -")
hash := zobrist.Hash(&pos)
data := board.Encode(&pos)

// Get stats with optional filtering.
filter := &query.Filter{
    TimeControlCategory: intPtr(2), // blitz only
}

view, err := query.PositionView(ctx, pool, hash, data, filter)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Position: %d games (W:%d D:%d B:%d)\n",
    view.Stats.Total, view.Stats.WhiteWins, view.Stats.Draws, view.Stats.BlackWins)

for _, mv := range view.Moves {
    fmt.Printf("  %s: %d games (W:%d D:%d B:%d)\n",
        mv.UCI, mv.Stats.Total, mv.Stats.WhiteWins, mv.Stats.Draws, mv.Stats.BlackWins)
}
```

## Design

Architecture and design decisions are documented in [`docs/design/`](docs/design/). Key documents:

| Doc | Topic |
|-----|-------|
| [001](docs/design/001-graph-model.md) | Graph model: positions as nodes, moves as edges |
| [002](docs/design/002-filtering-and-aggregation.md) | Filtering and aggregation at query time |
| [003](docs/design/003-positions-and-identity.md) | Position identity: Zobrist hash + encoded data |
| [008](docs/design/008-schema.md) | PostgreSQL schema (5 tables) |
| [010](docs/design/010-position-data-encoding.md) | 35-byte position encoding format |
| [011](docs/design/011-zobrist-hashing.md) | Custom Zobrist hash (not Polyglot) |
| [012](docs/design/012-implementation-plan.md) | Implementation plan |

Each code package also has a `DECISIONS.md` documenting implementation-level decisions.

## Testing

```bash
# Run all tests (sequential — required for database tests)
go test ./... -count=1 -p 1

# Run a specific package
go test ./chess/move/ -v

# Run perft validation (move generation correctness)
go test ./chess/move/ -v -run TestPerft
```

The chess library packages (`chess/*`) have no external dependencies and can be tested without a database. The `db`, `ingest`, and `query` packages require a running PostgreSQL instance.

### Test database

Tests expect a `chessbook_test` database on `localhost:5432` with user/password `postgres`. Override with the `CHESSBOOK_DATABASE_URL` environment variable:

```bash
CHESSBOOK_DATABASE_URL="postgres://user:pass@host:5432/mydb?sslmode=disable" go test ./... -p 1
```

## Key implementation details

### Chess960 support

The chess library supports Chess960 (Fischer Random) positions:

- **Castling rights** use a per-file bitmask (`[2]uint8` indexed by colour), supporting rooks on any file
- **FEN parsing** accepts both standard (`KQkq`) and Shredder-FEN (`AHah`) castling notation
- **Move generation** handles Chess960 castling edge cases (king/rook overlap on source/destination squares)

### Position identity

Two positions are considered identical if they have the same piece placement, side to move, castling rights, and en passant status. The halfmove clock and fullmove number are excluded.

**En passant normalisation**: The en passant square is only included in the position identity when a legal en passant capture actually exists. This prevents phantom EP squares from creating false position distinctions. Both the Zobrist hash and the 35-byte encoding apply this normalisation.

**Zobrist hashing**: Uses a custom 793-value random table generated from a fixed PCG seed. Polyglot hashing was considered but rejected because it cannot represent Chess960 castling rights (it has only 4 castling keys for KQkq, whereas Chess960 needs up to 16 for arbitrary rook files).

### Ingestion pipeline

1. **Parse**: Extract games from PGN (tags, SAN moves, result)
2. **Validate**: Replay every move to verify legality; reject entire file on any error
3. **Ingest**: Per-game transactions — upsert positions and edges, insert join table rows
4. **Dedup**: SHA-256 hash of Seven Tag Roster + movetext detects duplicate games

Games with `*` (unknown/ongoing) result are rejected during validation, as they would corrupt win/loss/draw statistics.

### Database schema

Five tables:

- **positions**: Zobrist hash (index) + encoded position data (unique key)
- **edges**: Source position → target position via UCI move string
- **games**: Player names, result, date, time control, dedup hash
- **position_games**: Join table (which games reached which positions)
- **edge_games**: Join table (which games played which moves)

Game IDs are stored in join tables rather than array columns to avoid O(n) array append costs on hot positions like 1.e4.
