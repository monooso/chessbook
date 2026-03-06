# 004: Game ID Format

Game IDs are auto-increment 64-bit integers (`BIGSERIAL` in Postgres).

## Why not UUIDs?

UUIDs solve a distributed coordination problem: generating unique IDs across multiple writers without a shared sequence. This system doesn't have that problem. Games are ingested from PGN files by a single process (or a small number of workers sharing a database). A Postgres sequence handles concurrent inserts without races — that's what sequences are for.

The cost of UUIDs, however, is real and compounds in this design:

1. **Storage on nodes and edges.** Game IDs aren't just in the games table. Every node and every edge carries a set of game IDs (see [001-graph-model](001-graph-model.md)). A mainline opening position might appear in millions of games. At 16 bytes per UUID vs 8 bytes per int64, the difference across all nodes and edges adds up to a significant multiple of total storage.

2. **Index size and cache efficiency.** Larger keys mean larger B-tree indexes, which means fewer index entries fit in memory. More cache misses means more disk I/O, which dominates query time in practice.

3. **Index fragmentation.** Random UUIDs (v4) scatter inserts across the B-tree, degrading write performance and fragmenting the index. Time-ordered UUIDs (v7/ULID) fix the ordering problem but are still 16 bytes.

## Why int64, not int32?

Int32 (4 bytes, ~2.1 billion values) would likely be sufficient and halves storage compared to int64. But int64 is the default `BIGSERIAL` in Postgres, costs only 4 extra bytes per ID, and removes any concern about eventually running out of space. It's cheap insurance.

## Trade-off: database merges

The one scenario where UUIDs help is merging two independently-built databases — IDs are globally unique, so there are no collisions. With auto-increment IDs, a merge requires remapping IDs in one of the databases.

This is an acceptable trade-off. Database merging is a rare, planned operation. Paying a permanent storage and performance cost on every node and edge to avoid a one-time migration step is not worthwhile.
