# 002: Filtering and Aggregation

All statistical queries (win rate, move impact, etc.) are computed at query time by filtering game IDs against the games table. This document explains how and why.

## The filtering model

Every node and edge carries a set of game IDs (see [001-graph-model](001-graph-model.md)). To compute stats for any entity, we:

1. Take its bag of game IDs
2. Query the games table with the active filters (time control, player, date range, etc.)
3. Count outcomes (white win / black win / draw) from the matching games

This means the same position can show different win rates for classical vs blitz, for a specific player vs the general population, or for games played this year vs all time. The graph structure doesn't change; only the filter applied to the game IDs changes.

## Win/loss/draw computation

For a given node (position) with filters applied:

1. Fetch the node's game IDs
2. `SELECT game_id, result FROM games WHERE game_id IN (...) AND <filters>`
3. Count the results

The viewer decides how to interpret "win" and "loss" based on whose perspective they're interested in. The graph itself stores results in neutral terms (white win / black win / draw), not relative to a viewer.

## Move impact

Move impact answers: "how does playing this move compare to the overall position?"

For a position with outgoing edges (available moves):

1. Compute filtered W/L/D for the **node** (the position overall)
2. Compute filtered W/L/D for each **edge** (each specific move)
3. Compare: an edge whose win rate exceeds the node's win rate is a move that performs above average in this position

This is a comparison between a set and its subsets. The edge's game IDs are always a subset of the node's game IDs, because every game that played move X from position A is also a game that reached position A.

### Why not compare position A to position B?

We considered computing move impact by comparing the destination position's stats to the source position's stats. This would mix in transpositions: position B's stats include games that arrived via completely different move orders. That muddies the signal.

The edge-vs-node comparison isolates the question to: "among games that reached this position, how did the ones that played this move fare compared to the whole group?"

## Query cost

For a single position view, the query workload is:

1. Fetch the node and all its outgoing edges (position + available moves)
2. Collect all game IDs from the node and edges into one set
3. One query: `SELECT game_id, result FROM games WHERE game_id IN (...) AND <filters>`
4. Partition results in application code: bucket each game ID's outcome by which node/edge it belongs to

This is one or two database queries regardless of how many moves are available. The IN clause size is bounded by the number of games that reached this position, which for most positions is modest. Mainline opening positions (e.g. after 1.e4) will have large game sets, but these are also the most cacheable.

## Why not pre-aggregate?

A classical game player may handle a complex closed position well, but consistently fluff it in blitz. This kind of insight requires filtering by time control, which means W/L/D counts must be computed per filter combination.

Pre-aggregating across every combination of filters (color × time control × player × date range × ...) creates a combinatorial explosion in storage and makes adding new filter dimensions a schema change. Computing at query time avoids both problems and keeps the system flexible.

For hot paths (popular positions, common filter combinations), caching at the application layer is straightforward and doesn't require structural changes to the data model.

## Caching (deferred)

The caching strategy is intentionally deferred. Without real data and working code, any caching decisions would be speculative. The design accommodates caching — game IDs on nodes/edges are stable, filter combinations are deterministic, and results are pure functions of (game IDs × filters) — but the specifics (what to cache, where, eviction policy, invalidation on new game ingestion) should be informed by actual query patterns and performance profiling.
