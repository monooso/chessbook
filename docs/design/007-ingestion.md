# 007: Ingestion

Ingestion is the process of turning a PGN file into graph entries. For each game, we replay the moves sequentially, upserting position nodes and appending the game ID to the relevant nodes and edges.

## Two-phase processing

Ingestion is two-phase: **validate first, then ingest.**

In the validation phase, we parse every game in the file and reject the entire file if any game is malformed (see "File validation" below). No database writes occur during validation. This catches problems early and ensures we never partially ingest a file that contains bad data.

In the ingestion phase, we replay each validated game and upsert its positions and moves into the graph. Each game is an atomic unit of work during this phase. If a game fails mid-ingestion (e.g., a database write error), we roll back only that game. There is no file-level transaction wrapping thousands of games — we do not want to discard thousands of successful imports because one game hit a transient database error.

In short: validation failures reject the file; ingestion failures reject the game.

## Duplicate detection

We detect and reject duplicate games to avoid inflating statistics. The deduplication key is a hash of the Seven Tag Roster (Event, Site, Date, Round, White, Black, Result) plus the full movetext. This handles edge cases like the same two players playing multiple games on the same date (the movetext or round number will differ), and is robust even for online blitz where metadata alone might collide.

Games with missing or incomplete headers are rejected at the file validation stage (see above), so we do not need to handle deduplication for games with insufficient metadata.

## File validation

File validation is the first phase of ingestion (see "Two-phase processing" above). We parse every game in the file and reject the entire file if any game is malformed. This includes:

- Illegal moves
- Truncated move lists
- Missing required headers

No database writes occur during validation. This is a correctness-first tool, not a browser. Malformed input is the caller's problem.

When a file is rejected, we return a structured error describing what went wrong: which game (by index), what the problem was, and enough context to locate it in the original file. No formal logging subsystem is needed up front, but the error information must be sufficient for diagnosis.

## Source tracking (deferred)

Source tracking — recording which ingestion operation produced each game, to enable bulk deletion by source — is deferred. The use case is real but not critical for v1, and retrofitting it later is straightforward: add an ingestion operations table, add a nullable foreign key column to games, and backfill as needed. Nothing in the current design makes this harder to add later.

## Concurrency

Ingestion is parallelised at the file level: multiple files can be ingested concurrently, each processed by an independent worker. Within a single file, games are processed sequentially.

The core ingestion function for a single game is simple and sequential — it takes a game and a reference to the graph, and returns a result. Parallelism lives in the orchestration layer above it. The only shared resource is the graph itself, and batch results are merged under a lock.

This gives us a natural scaling lever for large imports without introducing complexity into the per-game ingestion logic. It is straightforward to implement in any language and avoids fine-grained contention on hot positions (e.g. 1. e4, which appears in ~40% of all games).
