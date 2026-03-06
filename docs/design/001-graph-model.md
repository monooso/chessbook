# 001: Graph Model

The core data structure is a directed graph of chess positions connected by moves.

## Nodes: Positions

A node represents a single chess position. It contains:

- **Position key**: a unique identifier derived from the position (see [003-positions-and-identity](003-positions-and-identity.md))
- **Game IDs**: the set of games in which this position occurred

A node does not store win/loss/draw counts directly. These are computed at query time from the game IDs (see [002-filtering-and-aggregation](002-filtering-and-aggregation.md)).

## Edges: Moves

An edge represents a move played from one position to another. It connects a source node (the position before the move) to a destination node (the position after the move). It contains:

- **Move**: the move that was played, stored in UCI notation (see [006-move-notation](006-move-notation.md))
- **Game IDs**: the set of games in which this move was played from this position

The edge's game IDs are always a subset of the source node's game IDs. A game appears on the edge only if the player in that game actually chose this move in this position.

## Games

A game record stores metadata about a single game:

- Players (white, black)
- Result (white win, black win, draw)
- Time control
- Date
- Starting position (standard or Chess960 etc.)
- Any other metadata useful for filtering

Games are the single source of truth for all metadata. Nodes and edges reference games by ID but never duplicate game metadata.

## Why game IDs on both nodes and edges?

We initially considered storing game IDs only on edges, since every game that reaches a position must have arrived via some move (or be the starting position). However, computing a position's aggregate stats would then require collecting game IDs from all incoming edges plus accounting for games that start at this position. This is fragile and expensive.

Storing game IDs on the node directly is redundant but makes the model uniform: every entity in the graph (node or edge) carries a set of game IDs, and filtered stats can be computed against any of them in exactly the same way.

## Why no pre-aggregated win/loss/draw counts?

Pre-aggregated counts seem appealing for read performance but break down once filtering is introduced. A position's win rate differs by time control, by player, by date range, and by any combination of these. Pre-aggregating across every filter dimension creates a combinatorial explosion.

Instead, we store raw game IDs and compute stats at query time by joining against the games table with the active filters. This keeps the graph structure simple and makes all filters first-class citizens without schema changes.

## Relationship summary

```
Node (Position A) ---Edge (Move X)---> Node (Position B)
     |                    |
     +-- game IDs         +-- game IDs (subset of A's)
```

A node's game IDs: "these games reached this position."
An edge's game IDs: "these games played this move from this position."
