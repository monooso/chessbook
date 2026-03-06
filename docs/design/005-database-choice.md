# 005: Database Choice

PostgreSQL is the database for this system.

## Why not a graph database?

The data is a graph, but the workload is relational.

Graph databases (Neo4j, etc.) earn their keep on complex traversals: "find all paths between A and B," "what's reachable within 3 hops of X." The core query pattern here is simpler:

1. Look up one position by key
2. Fetch its outgoing edges (available moves)
3. Collect game IDs from the node and edges
4. Filter those game IDs against game metadata

Steps 1-2 are a single-hop lookup — a foreign key join. Steps 3-4 are relational filtering: `SELECT ... FROM games WHERE game_id IN (...) AND <filters>`. This is exactly what relational databases are optimized for.

## Why not a document database?

Document databases (MongoDB, etc.) could nest edges inside position documents, making step 1-2 a single document fetch. But step 4 — filtering game IDs against game metadata — requires joining across documents. This is the thing document databases are worst at, and it's the operation at the heart of every query in this system (see [002-filtering-and-aggregation](002-filtering-and-aggregation.md)).

## Why not a hybrid approach?

Running a graph database for positions alongside Postgres for games doubles the operational surface: two systems to deploy, monitor, back up, and keep consistent. The benefit would be faster graph traversals, but there are no complex traversals in this system. Single-position lookups with one level of outgoing edges are a `WHERE` clause, not a graph query.

## Why PostgreSQL specifically?

No exotic capabilities are required. PostgreSQL is a well-understood, well-supported relational database with:

- Mature indexing (B-tree, GIN for array columns, hash indexes)
- Join tables and B-tree indexes, which handle the game ID association workload well (see [008-schema](008-schema.md))
- A broad ecosystem for tooling, hosting, backups, and monitoring
- Extensive documentation and community support

The choice is deliberately conservative. Debugging a familiar database under load is dramatically easier than debugging an unfamiliar one, and nothing in this design requires capabilities that PostgreSQL lacks.
