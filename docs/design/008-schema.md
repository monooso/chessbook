# 008: Schema

How the conceptual graph model (see [001-graph-model](001-graph-model.md)) maps to PostgreSQL tables.

## Tables

### positions

A row per distinct position in the graph.

| Column | Type | Notes |
|---|---|---|
| `id` | `BIGSERIAL` | Primary key |
| `zobrist_hash` | `BIGINT` | Zobrist hash of the position, used for fast lookup |
| `position_data` | `BYTEA` | Full position (piece placement, active color, castling rights, en passant). Authoritative identity — the hash is an index, not a key (see [003](003-positions-and-identity.md)) |

**Indexes:**
- B-tree on `zobrist_hash` (lookup path)
- Unique constraint on `position_data` (prevents duplicate nodes)

### edges

A row per distinct move between two positions.

| Column | Type | Notes |
|---|---|---|
| `id` | `BIGSERIAL` | Primary key |
| `source_position_id` | `BIGINT` | FK → `positions.id` |
| `target_position_id` | `BIGINT` | FK → `positions.id` |
| `move` | `VARCHAR(5)` | UCI notation (e.g. `e2e4`, `e7e8q`) |

**Indexes:**
- B-tree on `source_position_id` (fetch outgoing edges for a position)
- Unique constraint on `(source_position_id, move)` — from a given position, a given move always leads to the same result

### games

A row per ingested game.

| Column | Type | Notes |
|---|---|---|
| `id` | `BIGSERIAL` | Primary key |
| `white` | `TEXT` | White player name |
| `black` | `TEXT` | Black player name |
| `result` | `SMALLINT` | 1 = white win, 0 = draw, -1 = black win |
| `year` | `SMALLINT` | Game year (nullable) |
| `month` | `SMALLINT` | Game month (nullable) |
| `day` | `SMALLINT` | Game day (nullable) |
| `time_control` | `TEXT` | Raw time control string from PGN, preserved for provenance |
| `time_control_category` | `SMALLINT` | Derived at ingestion: 1 = bullet, 2 = blitz, 3 = rapid, 4 = classical. Null if unparseable or missing |
| `starting_fen` | `TEXT` | Starting position FEN. Null for standard starting position |
| `dedup_hash` | `BYTEA` | Hash of Seven Tag Roster + movetext (see [007](007-ingestion.md)) |

**Indexes:**
- Unique constraint on `dedup_hash` (duplicate detection)
- B-tree on `white` and `black` (player filtering)
- B-tree on `(year, month, day)` (date range filtering)
- B-tree on `time_control_category` (time control filtering)

#### Partial dates

PGN dates use `??` for unknown components (e.g. `2024.01.??`, `2024.??.??`). Rather than discarding partial dates or inventing data, the date is stored as three separate nullable columns. This preserves all available information: a game dated `2024.01.??` has year=2024, month=1, day=NULL.

#### Time control categories

The `time_control_category` is derived from the raw time control string at ingestion time. For time controls in the standard `base+increment` format (both in seconds), estimated game duration is `base + 40 * increment` (assuming ~40 moves). The category cutoffs follow FIDE/Lichess conventions:

- **Bullet** (1): estimated duration < 3 minutes
- **Blitz** (2): 3–10 minutes
- **Rapid** (3): 10–60 minutes
- **Classical** (4): ≥ 60 minutes

Multi-stage time controls (e.g. `40/5400:1800`) are parsed by summing all stages. Unparseable or missing values produce NULL.

Additional metadata columns (ECO code, event, site, round, Elo ratings, etc.) can be added as filtering needs emerge. The schema is not closed.

### position_games

Join table linking positions to the games that reached them.

| Column | Type | Notes |
|---|---|---|
| `position_id` | `BIGINT` | FK → `positions.id` |
| `game_id` | `BIGINT` | FK → `games.id` |

**Primary key:** `(position_id, game_id)`

**Indexes:**
- The primary key index covers lookups by position (the dominant query path: "which games reached this position?")
- B-tree on `game_id` (reverse lookup: needed for game deletion)

### edge_games

Join table linking edges to the games that played them.

| Column | Type | Notes |
|---|---|---|
| `edge_id` | `BIGINT` | FK → `edges.id` |
| `game_id` | `BIGINT` | FK → `games.id` |

**Primary key:** `(edge_id, game_id)`

**Indexes:**
- The primary key index covers lookups by edge
- B-tree on `game_id` (reverse lookup: needed for game deletion)

## Why join tables, not array columns

The earlier design documents describe positions and edges as carrying "a set of game IDs." The natural Postgres representation would be integer array columns (`BIGINT[]`). We use join tables instead.

The problem with arrays is append cost. Postgres arrays are values, not collections — appending an element rewrites the entire array. For a position like 1.e4, which appears in millions of games, every new game that reaches this position rewrites a multi-megabyte array. The same position is also the most frequently touched during ingestion, compounding the problem.

Join tables avoid this entirely. Adding a game to a position is a single row insert — O(1) regardless of how many games already reference that position. This is the operation relational databases are optimized for.

The trade-off is storage overhead. Each join table row carries ~23 bytes of Postgres tuple header on top of 16 bytes of payload (two `BIGINT` foreign keys), compared to 8 bytes per element in an array. At target scale (~600M position_games rows, ~300M edge_games rows), the join tables use roughly 35 GB vs ~7 GB for equivalent arrays. This is a meaningful but manageable difference, and it buys us O(1) appends, standard B-tree indexing, and straightforward query patterns.

## Query pattern

The core query for a single position view (see [002-filtering-and-aggregation](002-filtering-and-aggregation.md)):

```sql
-- Step 1: Find the position
SELECT id FROM positions
WHERE zobrist_hash = $1 AND position_data = $2;

-- Step 2: Fetch outgoing edges
SELECT id, target_position_id, move FROM edges
WHERE source_position_id = $position_id;

-- Step 3: Fetch filtered game results for the position and all its edges
SELECT
    'position' AS source,
    NULL AS edge_id,
    g.result
FROM position_games pg
JOIN games g ON g.id = pg.game_id
WHERE pg.position_id = $position_id
  AND g.year >= $year_from          -- example filters; all optional
  AND g.time_control_category = $tc
  AND g.white = $player

UNION ALL

SELECT
    'edge' AS source,
    eg.edge_id,
    g.result
FROM edge_games eg
JOIN games g ON g.id = eg.game_id
WHERE eg.edge_id = ANY($edge_ids)
  AND g.year >= $year_from
  AND g.time_control_category = $tc
  AND g.white = $player;
```

Steps 1 and 2 are index lookups. Step 3 is a single query that returns all the data needed to compute win/loss/draw for the position and each outgoing move. Filters target indexed columns (`year`, `time_control_category`, `white`/`black`). The application partitions the results by source (position vs each edge) and counts outcomes.

## Ingestion write pattern

For each game, the ingestion process (see [007-ingestion](007-ingestion.md)):

1. Insert a row into `games`, receiving the new `game_id`
2. For each position in the game:
   - Upsert into `positions` (insert if new, otherwise get existing `id`)
   - Insert into `position_games`
3. For each move in the game:
   - Upsert into `edges` (insert if new, otherwise get existing `id`)
   - Insert into `edge_games`

All writes for a single game are wrapped in a transaction. On failure, the transaction rolls back and no partial data is left in the database.

Upserts on `positions` use the unique constraint on `position_data`. Upserts on `edges` use the unique constraint on `(source_position_id, move)`.

## What this schema does not cover

- **Caching.** Deferred per [002](002-filtering-and-aggregation.md).
- **Source tracking.** Deferred per [007](007-ingestion.md).
- **Position data encoding.** The binary format of `position_data` (how piece placement, castling rights, active color, and en passant are packed into bytes) is an implementation detail, not a schema concern. The schema stores it as opaque `BYTEA`.
- **Bulk ingestion optimizations.** The write pattern above describes the logical sequence. Batching inserts, using `COPY`, or other bulk-loading techniques are implementation-time optimizations that don't change the schema.
