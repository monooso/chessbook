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

The position key is derived from the FEN *excluding* the halfmove clock and fullmove number. Two positions that are identical in every way except that one is on move 10 and the other on move 25 are, for practical purposes, the same position. Including move counters would split what should be a single node into many, defeating transposition detection.

### Why the halfmove clock is excluded

The halfmove clock (used for the fifty-move rule) is path-dependent: it depends on the sequence of moves that led to the position, not the position itself. The same board state can have different halfmove clocks depending on which game reached it and by what route. Storing it on a position node would be incoherent — the node represents many games, each potentially with a different clock value.

This is fine because the graph is a derived view, not the canonical store. The original game records (PGN or equivalent) are the source of truth. If the halfmove clock is ever needed (e.g. to evaluate fifty-move-rule claims), it can be derived by replaying the relevant game's move list from its starting position — a trivial operation over a flat list of 40-80 moves. The graph is never walked backward to reconstruct this information.

### Zobrist hashing

The expected implementation approach for position keys is Zobrist hashing. A Zobrist hash assigns a random bitstring to each combination of (piece, square, color) and XORs them together to produce a single integer key for the position. Additional bitstrings cover castling rights, en passant, and active color.

This has two advantages over using FEN strings directly:

1. **Speed**: integer comparison and hashing is faster than string comparison.
2. **Incremental updates**: when processing a game move by move, the hash can be updated incrementally by XORing out the old state and XORing in the new state, rather than recomputing from scratch.

The hash excludes the halfmove clock and fullmove number, consistent with the position key derivation described above.

### Hash collisions

At large scale (millions of games, billions of distinct positions), Zobrist hash collisions become a real concern. With 64-bit hashes the birthday paradox puts collision probability in non-trivial territory around 2^32 (~4 billion) distinct positions.

The approach is the same one used by every hash-based data structure: **hash as index, verify with the full position.**

1. **The Zobrist hash is used for fast lookup**, narrowing the search to a small number of candidate nodes in O(1).
2. **Each node stores the full position data** — piece placement, active color, castling rights, and en passant square — compactly (32 bytes or less).
3. **On lookup, the full position is compared** to confirm identity. Two positions that happen to collide on the hash get separate nodes.

The hash is a performance optimization, not an identity. The full position data is the authoritative key. The storage overhead is modest: at a billion positions, the full position data adds roughly 32 GB, which is manageable.

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
