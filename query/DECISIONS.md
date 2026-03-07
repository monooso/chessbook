# query package: decisions

## Single combined query

Following the design doc, the stats query uses a single `UNION ALL` query that fetches results for both the position and all its outgoing edges. The application then partitions the results by source type (position vs edge) and counts outcomes. This minimises database round-trips.

## Dynamic filter clause

Filters are applied via dynamically constructed SQL conditions. Each filter field adds a condition and a parameter placeholder. This avoids the complexity of a query builder while remaining safe from SQL injection (all values are parameterised).

## Position lookup by hash and data

Positions are looked up using both the Zobrist hash (for fast index lookup) and the encoded position data (for correctness, since hashes can collide). This matches the design doc's two-step identity check.

## Test isolation

Database integration tests in `ingest` and `query` packages share a test database. They must be run with `-p 1` (sequential package execution) to avoid interference. Each test truncates all tables at the start.
