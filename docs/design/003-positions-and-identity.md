# 003: Positions and Identity

How positions are identified, how the graph handles perspective, and how nonstandard starting positions fit in.

## Position identity

A position is identified by a key derived from the FEN representation. The FEN encodes everything needed to distinguish one position from another:

- Piece placement
- Active color (whose turn it is)
- Castling availability
- En passant target square
- Halfmove clock and fullmove number

Two games that reach the same position via different move orders (transpositions) map to the same node in the graph. This is by design: the graph is a position graph, not a game tree.

### FEN as key vs derived key

The full FEN string can serve as the position key, but it includes the halfmove clock and fullmove number, which may or may not be desirable for deduplication. Two positions that are identical in every way except that one is on move 10 and the other on move 25 are, for practical purposes, the same position. The exact key derivation (full FEN vs FEN minus move counters) is an implementation decision that can be deferred, but the design should anticipate it.

### Zobrist hashing

The expected implementation approach for position keys is Zobrist hashing. A Zobrist hash assigns a random bitstring to each combination of (piece, square, color) and XORs them together to produce a single integer key for the position. Additional bitstrings cover castling rights, en passant, and active color.

This has two advantages over using FEN strings directly:

1. **Speed**: integer comparison and hashing is faster than string comparison.
2. **Incremental updates**: when processing a game move by move, the hash can be updated incrementally by XORing out the old state and XORing in the new state, rather than recomputing from scratch.

The choice of which position components to include in the hash (e.g. whether to include move counters) determines the deduplication behaviour, just as with FEN-based keys.

## Perspective neutrality

The graph does not distinguish between "my moves" and "opponent's moves." Nodes are positions, edges are moves. The active color is part of the position (encoded in the FEN), so a node implicitly knows whose turn it is, but this is a property of the position, not of the viewer.

### Why no player perspective in the graph?

A user may be reviewing their own games, studying a specific grandmaster, or exploring positions they've never played. Baking a player perspective into the graph structure would limit it to one use case. Instead, perspective is a concern of the UI and the query filters:

- "How do I do in this position?" → filter by player
- "How does Magnus do in this position?" → filter by player
- "How does everyone do in this position?" → no player filter

The graph supports all of these without structural changes.

## Terminal positions

A checkmate or stalemate position is a node with no outgoing edges. It may still have incoming edges (many games can end in the same checkmate pattern).

A position that exists in the graph but has no outgoing edges is not an error. It simply means no game in the dataset continued from that position, either because it's a terminal position (checkmate/stalemate) or because all recorded games in the dataset ended there (resignation, timeout, etc.).

A position that does not exist in the graph at all simply means no game in the dataset reached it. The application handles this the same as any empty query result.

## Nonstandard starting positions

Chess960 (and other variants with nonstandard starting positions) require no special treatment in the graph structure. A Chess960 game begins at a different starting position than a standard game, which means it enters the graph at a different starting node. From there, the graph works identically.

If a user wants to exclude Chess960 games, they use a game metadata filter. If they want to explore a specific Chess960 starting position, they navigate to that node. The graph doesn't need to know or care whether a position came from a standard or nonstandard game — a position is a position.

The starting position of each game is recorded in the game metadata (see [001-graph-model](001-graph-model.md)), which is what makes filtering by variant possible.
