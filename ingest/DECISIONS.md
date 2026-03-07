# ingest package: decisions

## Two-phase processing

Following the design doc, validation and ingestion are separate phases. `ValidateFile` replays all moves and rejects the entire file on any error. `IngestFile` then writes to the database game-by-game with per-game transactions.

## ValidatedGame carries replayed data

The `ValidatedGame` struct stores the full sequence of positions and moves produced during validation. This avoids replaying the game a second time during ingestion. The trade-off is memory — each validated game holds all its positions in memory simultaneously. This is acceptable because games are typically short (40–80 moves), and the entire file must be valid before ingestion begins.

## Dedup hash uses raw PGN data

The dedup hash is computed from the raw `pgn.Game` (Seven Tag Roster + SAN movetext), not from the validated moves. This matches the design doc requirement and ensures the hash is stable regardless of internal move representation. The `IngestFile` function takes both `validated` and `raw` slices for this reason.

## Upsert pattern

Position and edge upserts use `INSERT ... ON CONFLICT ... DO UPDATE ... RETURNING id`. The update is a no-op (sets a column to its existing value) — this is a common PostgreSQL pattern to always get the row ID back, whether the row was inserted or already existed.

## Join table ON CONFLICT DO NOTHING

`position_games` and `edge_games` inserts use `ON CONFLICT DO NOTHING` because a game might visit the same position multiple times (e.g., repetition). We only need one join row per (position, game) pair.

## Duplicate detection via PostgreSQL error code

Duplicate games are detected by catching PostgreSQL error code 23505 (unique_violation) on the `dedup_hash` constraint. This is simpler than a pre-check SELECT and handles concurrent ingestion correctly (no TOCTOU race).

## Time control category

Time control parsing follows the design doc: `base + 40 * increment` for estimated game duration, with standard category cutoffs (bullet < 3min, blitz 3–10min, rapid 10–60min, classical ≥ 60min). Multi-stage time controls (e.g., `40/5400:1800`) are supported by summing all stages. Unparseable values produce NULL.
